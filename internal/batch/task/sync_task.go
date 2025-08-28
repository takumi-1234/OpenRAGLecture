// open-rag-lecture/internal/batch/task/sync_task.go

package task

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/takumi-1234/OpenRAGLecture/internal/batch/processor"
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/repository"
	"gorm.io/gorm"
)

const (
	documentBatchSize = 10
	chunkBatchSize    = 100
)

// SyncTask handles the synchronization of documents to the vector database.
type SyncTask struct {
	db            *gorm.DB
	fileStorage   repository.FileStorage
	chunkProc     processor.ChunkProcessor
	embeddingRepo repository.EmbeddingRepository
	vectorRepo    repository.VectorRepository
}

// NewSyncTask creates a new SyncTask.
func NewSyncTask(
	db *gorm.DB,
	fileStorage repository.FileStorage,
	chunkProc processor.ChunkProcessor,
	embeddingRepo repository.EmbeddingRepository,
	vectorRepo repository.VectorRepository,
) *SyncTask {
	return &SyncTask{
		db:            db,
		fileStorage:   fileStorage,
		chunkProc:     chunkProc,
		embeddingRepo: embeddingRepo,
		vectorRepo:    vectorRepo,
	}
}

// Run executes the synchronization task.
func (t *SyncTask) Run(ctx context.Context) error {
	log.Println("Starting document synchronization task...")

	// At the beginning of the batch job, ensure the collection exists.
	if err := t.vectorRepo.EnsureCollectionExists(ctx); err != nil {
		return fmt.Errorf("failed to ensure Qdrant collection exists: %w", err)
	}

	for {
		docs, err := t.findUnprocessedDocuments(ctx)
		if err != nil {
			return fmt.Errorf("failed to find unprocessed documents: %w", err)
		}

		if len(docs) == 0 {
			log.Println("No new documents to process. Task finished.")
			return nil
		}

		log.Printf("Found %d new documents to process.", len(docs))

		for _, doc := range docs {
			if err := t.processDocument(ctx, doc); err != nil {
				log.Printf("ERROR: Failed to process document ID %d: %v", doc.ID, err)
				continue
			}
			log.Printf("Successfully processed document ID %d.", doc.ID)
		}

		if len(docs) < documentBatchSize {
			break
		}
	}

	log.Println("Document synchronization task completed.")
	return nil
}

// findUnprocessedDocuments queries the database for documents that don't have corresponding chunks.
func (t *SyncTask) findUnprocessedDocuments(ctx context.Context) ([]*model.Document, error) {
	var docs []*model.Document
	err := t.db.WithContext(ctx).
		Where("id NOT IN (SELECT DISTINCT document_id FROM chunks)").
		Limit(documentBatchSize).
		Find(&docs).Error
	return docs, err
}

// processDocument handles the full pipeline for a single document.
func (t *SyncTask) processDocument(ctx context.Context, doc *model.Document) error {
	fileContent, err := t.fileStorage.Get(ctx, doc.SourceURI)
	if err != nil {
		return fmt.Errorf("failed to get file from storage for doc %d: %w", doc.ID, err)
	}

	pages, chunks, err := t.chunkProc.Process(ctx, doc, fileContent)
	if err != nil {
		return fmt.Errorf("failed to chunk document %d: %w", doc.ID, err)
	}

	if len(chunks) == 0 {
		log.Printf("Document ID %d resulted in 0 chunks. Skipping.", doc.ID)
		return nil
	}

	for i := 0; i < len(chunks); i += chunkBatchSize {
		end := i + chunkBatchSize
		if end > len(chunks) {
			end = len(chunks)
		}
		chunkBatch := chunks[i:end]

		log.Printf("Processing chunk batch for doc %d: %d-%d of %d", doc.ID, i, end-1, len(chunks))

		texts := make([]string, len(chunkBatch))
		for j, c := range chunkBatch {
			texts[j] = c.Text
		}
		// ★★★ 修正点: taskTypeに "RETRIEVAL_DOCUMENT" を指定 ★★★
		vectors, err := t.embeddingRepo.CreateEmbeddings(ctx, texts, "RETRIEVAL_DOCUMENT")
		if err != nil {
			return fmt.Errorf("failed to create embeddings for doc %d: %w", doc.ID, err)
		}

		err = t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&pages).Error; err != nil {
				// This might fail if pages already exist, needs more robust logic
				// For now, we assume pages are new for each document.
			}

			if err := tx.Create(&chunkBatch).Error; err != nil {
				return err
			}

			if err := t.vectorRepo.Upsert(ctx, chunkBatch, vectors); err != nil {
				return fmt.Errorf("failed to upsert vectors for doc %d: %w", doc.ID, err)
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("database transaction failed for doc %d: %w", doc.ID, err)
		}

		time.Sleep(500 * time.Millisecond)
	}

	return nil
}
