// internal/tests/usecase/file_interactor_test.go
package usecase_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
	"github.com/takumi-1234/OpenRAGLecture/internal/tests/mocks"
	"github.com/takumi-1234/OpenRAGLecture/internal/usecase/interactor"
	appErrors "github.com/takumi-1234/OpenRAGLecture/pkg/errors"
)

func TestFileInteractor_Upload(t *testing.T) {
	ctx := context.Background()

	// --- 共通のテストデータをセットアップ ---
	courseID := uint64(101)
	semesterID := uint64(1)
	fileName := "lecture.pdf"
	fileContent := "dummy pdf content"
	course := &model.Course{
		Base:       model.Base{ID: courseID},
		SemesterID: semesterID,
	}
	savedPath := "101/some-uuid-lecture.pdf"

	// --- テストケースの実行 ---
	t.Run("Success_HappyPath", func(t *testing.T) {
		// Arrange
		mockDocRepo := new(mocks.MockDocumentRepository)
		mockFileStorage := new(mocks.MockFileStorage)
		mockCourseRepo := new(mocks.MockCourseRepository)
		fileInteractor := interactor.NewFileInteractor(mockDocRepo, mockFileStorage, mockCourseRepo)

		// ★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★
		// 修正点: io.Readerをサブテスト内で初期化
		// ★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★
		var file io.Reader = bytes.NewBufferString(fileContent)

		mockCourseRepo.On("FindByID", mock.Anything, courseID).Return(course, nil).Once()
		mockFileStorage.On("Save", mock.Anything, courseID, fileName, []byte(fileContent)).Return(savedPath, nil).Once()
		mockDocRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Document")).Return(nil).Once()

		// Act
		doc, err := fileInteractor.Upload(ctx, courseID, fileName, file)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, doc)
		assert.Equal(t, courseID, doc.CourseID)
		assert.Equal(t, semesterID, doc.SemesterID)
		assert.Equal(t, fileName, doc.Title)
		assert.Equal(t, savedPath, doc.SourceURI)
		mockCourseRepo.AssertExpectations(t)
		mockFileStorage.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("Failure_WhenCourseNotFound", func(t *testing.T) {
		// Arrange
		mockDocRepo := new(mocks.MockDocumentRepository)
		mockFileStorage := new(mocks.MockFileStorage)
		mockCourseRepo := new(mocks.MockCourseRepository)
		fileInteractor := interactor.NewFileInteractor(mockDocRepo, mockFileStorage, mockCourseRepo)
		var file io.Reader = bytes.NewBufferString(fileContent)

		mockCourseRepo.On("FindByID", mock.Anything, courseID).Return(nil, appErrors.ErrCourseNotFound).Once()

		// Act
		_, err := fileInteractor.Upload(ctx, courseID, fileName, file)

		// Assert
		assert.ErrorIs(t, err, appErrors.ErrCourseNotFound)
		mockFileStorage.AssertNotCalled(t, "Save")
		mockDocRepo.AssertNotCalled(t, "Create")
		mockCourseRepo.AssertExpectations(t)
	})

	t.Run("Failure_WhenFileStorageSaveFails", func(t *testing.T) {
		// Arrange
		mockDocRepo := new(mocks.MockDocumentRepository)
		mockFileStorage := new(mocks.MockFileStorage)
		mockCourseRepo := new(mocks.MockCourseRepository)
		fileInteractor := interactor.NewFileInteractor(mockDocRepo, mockFileStorage, mockCourseRepo)
		var file io.Reader = bytes.NewBufferString(fileContent)

		mockCourseRepo.On("FindByID", mock.Anything, courseID).Return(course, nil).Once()
		mockFileStorage.On("Save", mock.Anything, courseID, fileName, []byte(fileContent)).Return("", errors.New("disk full")).Once()

		// Act
		_, err := fileInteractor.Upload(ctx, courseID, fileName, file)

		// Assert
		assert.Error(t, err)
		assert.ErrorIs(t, err, appErrors.ErrFileUploadFailed)
		mockDocRepo.AssertNotCalled(t, "Create")
		mockCourseRepo.AssertExpectations(t)
		mockFileStorage.AssertExpectations(t)
	})

	t.Run("Failure_WhenDBCreateFails_ShouldCleanupFile", func(t *testing.T) {
		// Arrange
		mockDocRepo := new(mocks.MockDocumentRepository)
		mockFileStorage := new(mocks.MockFileStorage)
		mockCourseRepo := new(mocks.MockCourseRepository)
		fileInteractor := interactor.NewFileInteractor(mockDocRepo, mockFileStorage, mockCourseRepo)
		var file io.Reader = bytes.NewBufferString(fileContent)

		mockCourseRepo.On("FindByID", mock.Anything, courseID).Return(course, nil).Once()
		mockFileStorage.On("Save", mock.Anything, courseID, fileName, []byte(fileContent)).Return(savedPath, nil).Once()
		mockDocRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Document")).Return(errors.New("db error")).Once()
		mockFileStorage.On("Delete", mock.Anything, savedPath).Return(nil).Once()

		// Act
		_, err := fileInteractor.Upload(ctx, courseID, fileName, file)

		// Assert
		assert.ErrorIs(t, err, appErrors.ErrInternalServerError)
		mockCourseRepo.AssertExpectations(t)
		mockFileStorage.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})
}
