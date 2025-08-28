// internal/tests/handler/qa_handler_test.go
package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/takumi-1234/OpenRAGLecture/internal/interface/handler"
	"github.com/takumi-1234/OpenRAGLecture/internal/tests/mocks"
	"github.com/takumi-1234/OpenRAGLecture/internal/usecase/input"
	appErrors "github.com/takumi-1234/OpenRAGLecture/pkg/errors"
)

func TestQAHandler_Ask(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const testUserID = uint64(1)
	const testCourseID = uint64(101)

	askInput := input.AskInput{
		CourseID: testCourseID,
		Query:    "What is RAG?",
	}

	t.Run("Success_WhenEnrolled", func(t *testing.T) {
		// Arrange
		mockQAUsecase := new(mocks.MockQAUsecase)
		mockEnrollmentRepo := new(mocks.MockEnrollmentRepository)
		qaHandler := handler.NewQAHandler(mockQAUsecase, mockEnrollmentRepo)
		router := gin.New()
		router.POST("/api/qa/ask", authMiddlewareMock(testUserID), qaHandler.Ask)

		body, _ := json.Marshal(askInput)
		req, _ := http.NewRequest(http.MethodPost, "/api/qa/ask", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Mock dependencies
		mockEnrollmentRepo.On("IsEnrolled", mock.Anything, testUserID, testCourseID).Return(true, nil).Once()
		expectedAnswer := "This is the answer from the LLM."

		// We need to match the input struct with the UserID filled in.
		expectedUsecaseInput := askInput
		expectedUsecaseInput.UserID = testUserID
		mockQAUsecase.On("Ask", mock.Anything, expectedUsecaseInput).Return(expectedAnswer, nil).Once()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
		var respBody map[string]string
		_ = json.Unmarshal(rr.Body.Bytes(), &respBody)
		assert.Equal(t, expectedAnswer, respBody["answer"])
		mockEnrollmentRepo.AssertExpectations(t)
		mockQAUsecase.AssertExpectations(t)
	})

	t.Run("Failure_WhenNotEnrolled", func(t *testing.T) {
		// Arrange
		mockQAUsecase := new(mocks.MockQAUsecase)
		mockEnrollmentRepo := new(mocks.MockEnrollmentRepository)
		qaHandler := handler.NewQAHandler(mockQAUsecase, mockEnrollmentRepo)
		router := gin.New()
		router.POST("/api/qa/ask", authMiddlewareMock(testUserID), qaHandler.Ask)

		body, _ := json.Marshal(askInput)
		req, _ := http.NewRequest(http.MethodPost, "/api/qa/ask", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Mock dependencies
		mockEnrollmentRepo.On("IsEnrolled", mock.Anything, testUserID, testCourseID).Return(false, nil).Once()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, rr.Body.String(), appErrors.ErrNotEnrolled.Error())
		mockQAUsecase.AssertNotCalled(t, "Ask") // Usecase should not be called
		mockEnrollmentRepo.AssertExpectations(t)
	})

	t.Run("Failure_WhenEnrollmentCheckFails", func(t *testing.T) {
		// Arrange
		mockQAUsecase := new(mocks.MockQAUsecase)
		mockEnrollmentRepo := new(mocks.MockEnrollmentRepository)
		qaHandler := handler.NewQAHandler(mockQAUsecase, mockEnrollmentRepo)
		router := gin.New()
		router.POST("/api/qa/ask", authMiddlewareMock(testUserID), qaHandler.Ask)

		body, _ := json.Marshal(askInput)
		req, _ := http.NewRequest(http.MethodPost, "/api/qa/ask", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Mock dependencies
		mockEnrollmentRepo.On("IsEnrolled", mock.Anything, testUserID, testCourseID).Return(false, errors.New("db connection error")).Once()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		mockQAUsecase.AssertNotCalled(t, "Ask")
		mockEnrollmentRepo.AssertExpectations(t)
	})

	t.Run("Failure_BadRequest_InvalidJSON", func(t *testing.T) {
		// Arrange
		mockQAUsecase := new(mocks.MockQAUsecase)
		mockEnrollmentRepo := new(mocks.MockEnrollmentRepository)
		qaHandler := handler.NewQAHandler(mockQAUsecase, mockEnrollmentRepo)
		router := gin.New()
		router.POST("/api/qa/ask", authMiddlewareMock(testUserID), qaHandler.Ask)

		// Invalid JSON body
		req, _ := http.NewRequest(http.MethodPost, "/api/qa/ask", bytes.NewBufferString(`{"course_id":101, "query":`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mockEnrollmentRepo.AssertNotCalled(t, "IsEnrolled")
		mockQAUsecase.AssertNotCalled(t, "Ask")
	})
}
