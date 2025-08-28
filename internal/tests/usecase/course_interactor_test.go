// internal/tests/usecase/course_interactor_test.go
package usecase_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
	"github.com/takumi-1234/OpenRAGLecture/internal/tests/mocks"
	"github.com/takumi-1234/OpenRAGLecture/internal/usecase/interactor"
	appErrors "github.com/takumi-1234/OpenRAGLecture/pkg/errors"
)

func TestCourseInteractor_EnrollUser(t *testing.T) {
	ctx := context.Background()
	mockCourseRepo := new(mocks.MockCourseRepository)
	mockEnrollmentRepo := new(mocks.MockEnrollmentRepository)
	courseInteractor := interactor.NewCourseInteractor(mockCourseRepo, mockEnrollmentRepo)

	userID := uint64(1)
	courseID := uint64(101)

	t.Run("Success_WhenNotEnrolled", func(t *testing.T) {
		// Arrange
		mockCourseRepo.On("FindByID", ctx, courseID).Return(&model.Course{Base: model.Base{ID: courseID}}, nil).Once()
		mockEnrollmentRepo.On("IsEnrolled", ctx, userID, courseID).Return(false, nil).Once()
		mockEnrollmentRepo.On("Create", ctx, mock.AnythingOfType("*model.Enrollment")).Return(nil).Once()

		// Act
		err := courseInteractor.EnrollUser(ctx, userID, courseID)

		// Assert
		assert.NoError(t, err)
		mockCourseRepo.AssertExpectations(t)
		mockEnrollmentRepo.AssertExpectations(t)
	})

	t.Run("Failure_WhenCourseNotFound", func(t *testing.T) {
		// Arrange
		mockCourseRepo.On("FindByID", ctx, courseID).Return(nil, appErrors.ErrCourseNotFound).Once()

		// Act
		err := courseInteractor.EnrollUser(ctx, userID, courseID)

		// Assert
		assert.ErrorIs(t, err, appErrors.ErrNotFound)
		mockCourseRepo.AssertExpectations(t)
		// 後続の処理が呼ばれないことを確認
		mockEnrollmentRepo.AssertNotCalled(t, "IsEnrolled")
		mockEnrollmentRepo.AssertNotCalled(t, "Create")
	})

	t.Run("Failure_AlreadyEnrolled_ReturnsConflictError", func(t *testing.T) {
		// Arrange
		mockCourseRepo.On("FindByID", ctx, courseID).Return(&model.Course{Base: model.Base{ID: courseID}}, nil).Once()
		mockEnrollmentRepo.On("IsEnrolled", ctx, userID, courseID).Return(true, nil).Once()

		// Act
		err := courseInteractor.EnrollUser(ctx, userID, courseID)

		// Assert
		assert.ErrorIs(t, err, appErrors.ErrConflict)
		mockCourseRepo.AssertExpectations(t)
		mockEnrollmentRepo.AssertExpectations(t)
		// 登録処理が呼ばれないことを確認
		mockEnrollmentRepo.AssertNotCalled(t, "Create")
	})
}
