package google

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/repository"
	"github.com/takumi-1234/OpenRAGLecture/pkg/config"

	"google.golang.org/genai"
)

type googleLLMRepository struct {
	client    *genai.Client
	modelName string
}

func NewGoogleLLMRepository(cfg config.GoogleConfig) (repository.LLMRepository, error) {
	ctx := context.Background()

	// ClientConfig に Project / Location / Backend を渡す（VertexAI を使う場合）
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  cfg.ProjectID,
		Location: cfg.Location,
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		log.Printf("Failed to create genai client for vertex ai. Project: %s, Location: %s", cfg.ProjectID, cfg.Location)
		return nil, fmt.Errorf("failed to create google genai client for vertex ai: %w", err)
	}

	return &googleLLMRepository{
		client:    client,
		modelName: cfg.LLMModel,
	}, nil
}

func (r *googleLLMRepository) GenerateContent(ctx context.Context, params repository.GenerateContentParams) (string, error) {
	// 結合プロンプトを作成（システム指示、コンテキスト、ユーザープロンプト）
	promptParts := make([]string, 0, 3)
	if params.SystemPrompt != "" {
		promptParts = append(promptParts, "[System]\n"+params.SystemPrompt)
	}

	contextText := buildContextString(params.ContextChunks)
	if contextText != "" {
		promptParts = append(promptParts, "[Context]\n"+contextText)
	}

	if params.UserPrompt != "" {
		promptParts = append(promptParts, "[User]\n"+params.UserPrompt)
	}

	combinedPrompt := strings.Join(promptParts, "\n\n")

	// GenerateContent を呼ぶ（公式サンプルに合わせる）
	res, err := r.client.Models.GenerateContent(ctx, r.modelName, genai.Text(combinedPrompt), nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if res == nil {
		return "", errors.New("no response from model")
	}

	// 生成テキストを返す（res.Text() は公式サンプルのパターン）
	return res.Text(), nil
}

func (r *googleLLMRepository) GenerateContentStream(ctx context.Context, params repository.GenerateContentParams, writer io.Writer) error {
	// 結合プロンプトを作成
	promptParts := make([]string, 0, 3)
	if params.SystemPrompt != "" {
		promptParts = append(promptParts, "[System]\n"+params.SystemPrompt)
	}

	contextText := buildContextString(params.ContextChunks)
	if contextText != "" {
		promptParts = append(promptParts, "[Context]\n"+contextText)
	}

	if params.UserPrompt != "" {
		promptParts = append(promptParts, "[User]\n"+params.UserPrompt)
	}

	combinedPrompt := strings.Join(promptParts, "\n\n")

	// ストリーミング呼び出し（公式サンプルに準拠）
	iter := r.client.Models.GenerateContentStream(ctx, r.modelName, genai.Text(combinedPrompt), nil)
	for result, err := range iter {
		if err != nil {
			return fmt.Errorf("chat stream error: %w", err)
		}
		text := result.Text() // または result.Candidates... など、どの形で取得するかに応じて適宜
		if _, wErr := writer.Write([]byte(text)); wErr != nil {
			log.Printf("write error: %v", wErr)
			return wErr
		}
	}

	return nil
}

func buildContextString(chunks []model.RetrievedChunk) string {
	if len(chunks) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("--- Context Information ---\n")
	for i, chunk := range chunks {
		sb.WriteString(fmt.Sprintf("\n[Reference %d]\n%s\n", i+1, chunk.Chunk.Text))
	}
	sb.WriteString("--- End of Context ---\n")
	return sb.String()
}
