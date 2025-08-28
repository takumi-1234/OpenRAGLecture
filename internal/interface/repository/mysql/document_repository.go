// OpenRAGLecture/internal/interface/repository/mysql/document_repository.go
package mysql

import (
	"context"
	"errors"
	"fmt"

	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/repository"
	appErrors "github.com/takumi-1234/OpenRAGLecture/pkg/errors"
	"gorm.io/gorm"
)

type documentRepository struct {
	db *gorm.DB
}

// NewDocumentRepository creates a new DocumentRepository implementation.
func NewDocumentRepository(db *gorm.DB) repository.DocumentRepository {
	return &documentRepository{db: db}
}

func (r *documentRepository) FindByID(ctx context.Context, id uint64) (*model.Document, error) {
	var doc model.Document
	if err := r.db.WithContext(ctx).First(&doc, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrDocumentNotFound
		}
		return nil, err
	}
	return &doc, nil
}

func (r *documentRepository) Create(ctx context.Context, doc *model.Document) error {
	return r.db.WithContext(ctx).Create(doc).Error
}

// FullTextSearch performs a natural language full-text search.
func (r *documentRepository) FullTextSearch(ctx context.Context, query string, courseID uint64, limit int) ([]model.RetrievedChunk, error) {
	var results []struct {
		model.Chunk
		Score float32
	}

	// This subquery finds the relevant pages using FULLTEXT index.
	// Then we join with chunks associated with those pages.
	sql := `
		SELECT c.*, p_score.score
		FROM chunks c
		INNER JOIN (
			SELECT p.id as page_id, MATCH(p.text) AGAINST(? IN NATURAL LANGUAGE MODE) as score
			FROM pages p
			INNER JOIN documents d ON p.document_id = d.id
			WHERE d.course_id = ? AND MATCH(p.text) AGAINST(? IN NATURAL LANGUAGE MODE) > 0
		) as p_score ON c.page_id = p_score.page_id
		ORDER BY p_score.score DESC
		LIMIT ?;
	`
	err := r.db.WithContext(ctx).Raw(sql, query, courseID, query, limit).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("full-text search failed: %w", err)
	}

	retrievedChunks := make([]model.RetrievedChunk, len(results))
	for i, res := range results {
		retrievedChunks[i] = model.RetrievedChunk{
			Chunk: res.Chunk,
			Score: res.Score,
		}
	}

	return retrievedChunks, nil
}
