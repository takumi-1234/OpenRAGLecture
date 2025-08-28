// internal/tests/handler/file_handler_test.go
package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
	"github.com/takumi-1234/OpenRAGLecture/internal/interface/handler"
	"github.com/takumi-1234/OpenRAGLecture/internal/tests/mocks"
)

func TestFileHandler_Upload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const courseID = 101

	// Helper function to create a multipart form request
	createMultipartRequest := func(fileContent, courseIDStr string) (*http.Request, string) {
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)

		if fileContent != "" {
			part, _ := writer.CreateFormFile("file", "test.pdf")
			_, _ = io.WriteString(part, fileContent)
		}

		if courseIDStr != "" {
			_ = writer.WriteField("course_id", courseIDStr)
		}

		writer.Close()

		req, _ := http.NewRequest(http.MethodPost, "/api/files/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		return req, writer.FormDataContentType()
	}

	t.Run("Success_HappyPath", func(t *testing.T) {
		// Arrange
		mockFileUsecase := new(mocks.MockFileUsecase)
		fileHandler := handler.NewFileHandler(mockFileUsecase)
		router := gin.New()
		router.POST("/api/files/upload", fileHandler.Upload)

		req, _ := createMultipartRequest("dummy pdf content", strconv.Itoa(courseID))
		rr := httptest.NewRecorder()

		// Mock Usecase response
		mockFileUsecase.On("Upload", mock.Anything, uint64(courseID), "test.pdf", mock.Anything).Return(&model.Document{Base: model.Base{ID: 1}}, nil).Once()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusCreated, rr.Code)
		var respBody map[string]interface{}
		_ = json.Unmarshal(rr.Body.Bytes(), &respBody)
		assert.Equal(t, "File uploaded successfully. Processing has started.", respBody["message"])
		// JSON numbers are decoded as float64 by default
		assert.EqualValues(t, 1, respBody["document_id"])
		mockFileUsecase.AssertExpectations(t)
	})

	t.Run("Failure_InvalidCourseID_NotANumber", func(t *testing.T) {
		// Arrange
		mockFileUsecase := new(mocks.MockFileUsecase)
		fileHandler := handler.NewFileHandler(mockFileUsecase)
		router := gin.New()
		router.POST("/api/files/upload", fileHandler.Upload)

		req, _ := createMultipartRequest("dummy content", "invalid-id")
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid course_id")
		mockFileUsecase.AssertNotCalled(t, "Upload")
	})

	t.Run("Failure_MissingCourseID", func(t *testing.T) {
		// Arrange
		mockFileUsecase := new(mocks.MockFileUsecase)
		fileHandler := handler.NewFileHandler(mockFileUsecase)
		router := gin.New()
		router.POST("/api/files/upload", fileHandler.Upload)

		req, _ := createMultipartRequest("dummy content", "") // Missing course_id
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid course_id")
		mockFileUsecase.AssertNotCalled(t, "Upload")
	})

	t.Run("Failure_MissingFile", func(t *testing.T) {
		// Arrange
		mockFileUsecase := new(mocks.MockFileUsecase)
		fileHandler := handler.NewFileHandler(mockFileUsecase)
		router := gin.New()
		router.POST("/api/files/upload", fileHandler.Upload)

		req, _ := createMultipartRequest("", strconv.Itoa(courseID)) // Missing file
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "File not provided")
		mockFileUsecase.AssertNotCalled(t, "Upload")
	})

	t.Run("Failure_UsecaseReturnsError", func(t *testing.T) {
		// Arrange
		mockFileUsecase := new(mocks.MockFileUsecase)
		fileHandler := handler.NewFileHandler(mockFileUsecase)
		router := gin.New()
		router.POST("/api/files/upload", fileHandler.Upload)

		req, _ := createMultipartRequest("dummy content", strconv.Itoa(courseID))
		rr := httptest.NewRecorder()

		// Mock Usecase to return a generic error
		mockFileUsecase.On("Upload", mock.Anything, uint64(courseID), "test.pdf", mock.Anything).Return(nil, errors.New("usecase failed")).Once()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "usecase failed")
		mockFileUsecase.AssertExpectations(t)
	})
}
