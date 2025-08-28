// open-rag-lecture/internal/domain/repository/vector_repository.go
package repository

import (
	"context"

	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
)

type VectorRepository interface {
	// Upsert inserts or updates vectors (chunks) into the vector database.
	Upsert(ctx context.Context, chunks []*model.Chunk, vectors [][]float32) error
	// Search finds similar vectors based on a query vector.
	Search(ctx context.Context, queryVector []float32, courseID uint64, limit int) ([]model.RetrievedChunk, error)
	// RecreateCollection deletes a collection if it exists and creates a new one.
	// This is useful for ensuring a clean state, especially for testing.
	RecreateCollection(ctx context.Context) error
	// EnsureCollectionExists checks if a collection exists, and creates it if it does not.
	// This is useful for application startup or batch jobs.
	EnsureCollectionExists(ctx context.Context) error
}
