// OpenRAGLecture/internal/domain/repository/feedback_repository.go
package repository

import (
	"context"

	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
)

type FeedbackRepository interface {
	Create(ctx context.Context, feedback *model.Feedback) error
}
