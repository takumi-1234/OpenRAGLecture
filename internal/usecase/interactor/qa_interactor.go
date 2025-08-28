// OpenRAGLecture/internal/usecase/interactor/qa_interactor.go
package interactor

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/repository"
	"github.com/takumi-1234/OpenRAGLecture/internal/usecase/input"
	"github.com/takumi-1234/OpenRAGLecture/internal/usecase/port"
	"golang.org/x/sync/errgroup"
)

const (
	hybridSearchTopK     = 5
	rerankTopN           = 3
	systemPromptTemplate = `You are an excellent AI assistant for university lectures.
Please answer the user's question based ONLY on the provided context information below.
If the context does not contain the answer, state that you cannot answer based on the provided materials.
Do not make up information. Be concise, helpful, and accurate.

---
%s
---
`
)

type qaInteractor struct {
	docRepo       repository.DocumentRepository
	vectorRepo    repository.VectorRepository
	embeddingRepo repository.EmbeddingRepository
	llmRepo       repository.LLMRepository
	// TODO: Add cacheRepo and other dependencies
}

// NewQAInteractor creates a new instance of QAUsecase.
func NewQAInteractor(
	docRepo repository.DocumentRepository,
	vectorRepo repository.VectorRepository,
	embeddingRepo repository.EmbeddingRepository,
	llmRepo repository.LLMRepository,
) port.QAUsecase {
	return &qaInteractor{
		docRepo:       docRepo,
		vectorRepo:    vectorRepo,
		embeddingRepo: embeddingRepo,
		llmRepo:       llmRepo,
	}
}

func (i *qaInteractor) Ask(ctx context.Context, in input.AskInput) (string, error) {
	// 1. Create query embedding
	// ★★★ 修正点: taskTypeに "RETRIEVAL_QUERY" を指定 ★★★
	queryEmbeddings, err := i.embeddingRepo.CreateEmbeddings(ctx, []string{in.Query}, "RETRIEVAL_QUERY")
	if err != nil || len(queryEmbeddings) == 0 {
		return "", fmt.Errorf("failed to create query embedding: %w", err)
	}
	queryVector := queryEmbeddings[0]

	// 2. Hybrid Search (BM25 + Vector) in parallel
	var bm25Results, vectorResults []model.RetrievedChunk
	eg, gCtx := errgroup.WithContext(ctx)

	// BM25 (Full-text) search
	eg.Go(func() error {
		var err error
		bm25Results, err = i.docRepo.FullTextSearch(gCtx, in.Query, in.CourseID, hybridSearchTopK)
		if err != nil {
			return fmt.Errorf("BM25 search failed: %w", err)
		}
		return nil
	})

	// Vector search
	eg.Go(func() error {
		var err error
		vectorResults, err = i.vectorRepo.Search(gCtx, queryVector, in.CourseID, hybridSearchTopK)
		if err != nil {
			return fmt.Errorf("vector search failed: %w", err)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return "", err
	}

	// 3. Rerank/Merge results (using Reciprocal Rank Fusion - RRF)
	rerankedChunks := i.rerank(bm25Results, vectorResults, rerankTopN)
	if len(rerankedChunks) == 0 {
		return "I could not find any relevant information in the provided materials to answer your question.", nil
	}

	// 4. Generate response using LLM
	contextStr := i.buildContextString(rerankedChunks)
	finalPrompt := fmt.Sprintf(systemPromptTemplate, contextStr)

	params := repository.GenerateContentParams{
		SystemPrompt:  finalPrompt,
		UserPrompt:    in.Query,
		ContextChunks: rerankedChunks,
	}

	response, err := i.llmRepo.GenerateContent(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	// TODO: Save question, answer, sources to DB asynchronously

	return response, nil
}

func (i *qaInteractor) AskStream(ctx context.Context, in input.AskInput, writer io.Writer) error {
	// Similar logic as Ask, but calls GenerateContentStream
	// (Implementation is omitted for brevity but would follow the same RAG pipeline)
	// For this example, we'll just show the final LLM call part.

	// (Steps 1-3 would be identical to the Ask method)
	rerankedChunks := []model.RetrievedChunk{} // Assume this is populated

	contextStr := i.buildContextString(rerankedChunks)
	finalPrompt := fmt.Sprintf(systemPromptTemplate, contextStr)

	params := repository.GenerateContentParams{
		SystemPrompt:  finalPrompt,
		UserPrompt:    in.Query,
		ContextChunks: rerankedChunks,
	}

	return i.llmRepo.GenerateContentStream(ctx, params, writer)
}

// rerank combines and ranks search results using Reciprocal Rank Fusion (RRF).
// A simple k=60 is used for the ranking formula.
func (i *qaInteractor) rerank(listA, listB []model.RetrievedChunk, topN int) []model.RetrievedChunk {
	const k = 60.0
	scores := make(map[uint64]float64)
	chunkMap := make(map[uint64]model.RetrievedChunk)

	for rank, chunk := range listA {
		id := chunk.Chunk.ID
		scores[id] += 1.0 / (float64(rank) + k)
		if _, ok := chunkMap[id]; !ok {
			chunkMap[id] = chunk
		}
	}
	for rank, chunk := range listB {
		id := chunk.Chunk.ID
		scores[id] += 1.0 / (float64(rank) + k)
		if _, ok := chunkMap[id]; !ok {
			chunkMap[id] = chunk
		}
	}

	type rankedResult struct {
		ID    uint64
		Score float64
	}

	var ranked []rankedResult
	for id, score := range scores {
		ranked = append(ranked, rankedResult{ID: id, Score: score})
	}

	// Sort by score descending
	// (Using sort.Slice for production is better)
	for i := 0; i < len(ranked); i++ {
		for j := i + 1; j < len(ranked); j++ {
			if ranked[i].Score < ranked[j].Score {
				ranked[i], ranked[j] = ranked[j], ranked[i]
			}
		}
	}

	var finalChunks []model.RetrievedChunk
	for i := 0; i < len(ranked) && i < topN; i++ {
		finalChunks = append(finalChunks, chunkMap[ranked[i].ID])
	}

	return finalChunks
}

// buildContextString creates a single string from the context chunks.
func (i *qaInteractor) buildContextString(chunks []model.RetrievedChunk) string {
	var sb strings.Builder
	for i, chunk := range chunks {
		sb.WriteString(fmt.Sprintf("---\nDocument Snippet %d---\n", i+1))
		sb.WriteString(chunk.Chunk.Text)
		sb.WriteString("\n\n")
	}
	return sb.String()
}
