// open-rag-lecture/internal/usecase/interactor/course_interactor.go
package interactor

import (
	"context"
	"errors"

	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/repository"
	"github.com/takumi-1234/OpenRAGLecture/internal/usecase/port"
	appErrors "github.com/takumi-1234/OpenRAGLecture/pkg/errors"
)

type courseInteractor struct {
	courseRepo     repository.CourseRepository
	enrollmentRepo repository.EnrollmentRepository
}

// NewCourseInteractor creates a new instance of CourseUsecase.
func NewCourseInteractor(
	courseRepo repository.CourseRepository,
	enrollmentRepo repository.EnrollmentRepository,
) port.CourseUsecase {
	return &courseInteractor{
		courseRepo:     courseRepo,
		enrollmentRepo: enrollmentRepo,
	}
}

// EnrollUser handles the business logic for enrolling a user in a course.
func (i *courseInteractor) EnrollUser(ctx context.Context, userID, courseID uint64) error {
	// 1. Check if the course exists.
	_, err := i.courseRepo.FindByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, appErrors.ErrCourseNotFound) {
			return appErrors.ErrNotFound // Return a generic NotFound for the handler
		}
		return appErrors.ErrInternalServerError
	}

	// 2. Check if the user is already enrolled.
	isEnrolled, err := i.enrollmentRepo.IsEnrolled(ctx, userID, courseID)
	if err != nil {
		return appErrors.ErrInternalServerError
	}
	if isEnrolled {
		return appErrors.ErrConflict // User is already enrolled
	}

	// 3. Create the enrollment record.
	enrollment := &model.Enrollment{
		UserID:   userID,
		CourseID: courseID,
		Role:     "student", // Default role
	}

	if err := i.enrollmentRepo.Create(ctx, enrollment); err != nil {
		return appErrors.ErrInternalServerError
	}

	return nil
}
