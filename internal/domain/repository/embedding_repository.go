// OpenRAGLecture/internal/domain/repository/embedding_repository.go
package repository

import "context"

type EmbeddingRepository interface {
	// CreateEmbeddings generates vector embeddings for a batch of texts.
	// taskType is crucial for getting high-quality embeddings.
	// Use "RETRIEVAL_QUERY" for user queries and "RETRIEVAL_DOCUMENT" for documents.
	CreateEmbeddings(ctx context.Context, texts []string, taskType string) ([][]float32, error)
}
