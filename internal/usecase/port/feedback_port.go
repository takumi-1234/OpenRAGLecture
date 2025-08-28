// OpenRAGLecture/internal/usecase/port/feedback_port.go
package port

import (
	"context"

	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
)

// FeedbackUsecase defines the interface for handling user feedback.
type FeedbackUsecase interface {
	Submit(ctx context.Context, userID, answerID uint64, thumbsUp *bool, comment string) (*model.Feedback, error)
}
