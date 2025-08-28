// OpenRAGLecture/internal/usecase/interactor/feedback_interactor.go
package interactor

import (
	"context"

	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/repository"
	"github.com/takumi-1234/OpenRAGLecture/internal/usecase/port"
	appErrors "github.com/takumi-1234/OpenRAGLecture/pkg/errors"
)

type feedbackInteractor struct {
	feedbackRepo repository.FeedbackRepository
}

// NewFeedbackInteractor creates a new instance of FeedbackUsecase.
func NewFeedbackInteractor(feedbackRepo repository.FeedbackRepository) port.FeedbackUsecase {
	return &feedbackInteractor{
		feedbackRepo: feedbackRepo,
	}
}

func (i *feedbackInteractor) Submit(ctx context.Context, userID, answerID uint64, thumbsUp *bool, comment string) (*model.Feedback, error) {
	feedback := &model.Feedback{
		UserID:   userID,
		AnswerID: answerID,
		ThumbsUp: thumbsUp,
		Comment:  comment,
	}

	if err := i.feedbackRepo.Create(ctx, feedback); err != nil {
		return nil, appErrors.ErrInternalServerError
	}

	return feedback, nil
}
