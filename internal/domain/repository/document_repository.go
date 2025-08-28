// OpenRAGLecture/internal/domain/repository/document_repository.go
package repository

import (
	"context"

	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
)

type DocumentRepository interface {
	FindByID(ctx context.Context, id uint64) (*model.Document, error)
	Create(ctx context.Context, doc *model.Document) error
	// FullTextSearch performs a BM25-like search on the `pages` table.
	FullTextSearch(ctx context.Context, query string, courseID uint64, limit int) ([]model.RetrievedChunk, error)
}
