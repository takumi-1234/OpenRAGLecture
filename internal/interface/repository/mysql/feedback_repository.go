// OpenRAGLecture/internal/interface/repository/mysql/feedback_repository.go
package mysql

import (
	"context"

	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/repository"
	"gorm.io/gorm"
)

type feedbackRepository struct {
	db *gorm.DB
}

// NewFeedbackRepository creates a new FeedbackRepository implementation.
func NewFeedbackRepository(db *gorm.DB) repository.FeedbackRepository {
	return &feedbackRepository{db: db}
}

func (r *feedbackRepository) Create(ctx context.Context, feedback *model.Feedback) error {
	return r.db.WithContext(ctx).Create(feedback).Error
}
