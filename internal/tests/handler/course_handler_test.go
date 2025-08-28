// internal/tests/handler/course_handler_test.go
package handler_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/takumi-1234/OpenRAGLecture/internal/interface/handler"
	"github.com/takumi-1234/OpenRAGLecture/internal/tests/mocks"
	appErrors "github.com/takumi-1234/OpenRAGLecture/pkg/errors"
)

func TestCourseHandler_Enroll(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCourseUsecase := new(mocks.MockCourseUsecase)
	courseHandler := handler.NewCourseHandler(mockCourseUsecase)

	const testUserID = uint64(1)
	const testCourseID = uint64(101)

	router := gin.New()
	router.POST("/courses/:course_id/enrollments", authMiddlewareMock(testUserID), courseHandler.Enroll)

	t.Run("Success", func(t *testing.T) {
		mockCourseUsecase.On("EnrollUser", mock.Anything, testUserID, testCourseID).Return(nil).Once()

		url := fmt.Sprintf("/courses/%d/enrollments", testCourseID)
		req, _ := http.NewRequest(http.MethodPost, url, nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		mockCourseUsecase.AssertExpectations(t)
	})

	t.Run("Failure_CourseNotFound", func(t *testing.T) {
		mockCourseUsecase.On("EnrollUser", mock.Anything, testUserID, testCourseID).Return(appErrors.ErrNotFound).Once()

		url := fmt.Sprintf("/courses/%d/enrollments", testCourseID)
		req, _ := http.NewRequest(http.MethodPost, url, nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockCourseUsecase.AssertExpectations(t)
	})

	t.Run("Failure_AlreadyEnrolled_Conflict", func(t *testing.T) {
		mockCourseUsecase.On("EnrollUser", mock.Anything, testUserID, testCourseID).Return(appErrors.ErrConflict).Once()

		url := fmt.Sprintf("/courses/%d/enrollments", testCourseID)
		req, _ := http.NewRequest(http.MethodPost, url, nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusConflict, rr.Code)
		mockCourseUsecase.AssertExpectations(t)
	})

	t.Run("Failure_InvalidCourseID", func(t *testing.T) {
		url := "/courses/invalid/enrollments"
		req, _ := http.NewRequest(http.MethodPost, url, nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mockCourseUsecase.AssertNotCalled(t, "EnrollUser")
	})
}
