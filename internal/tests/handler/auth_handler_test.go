// internal/tests/handler/auth_handler_test.go
package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/takumi-1234/OpenRAGLecture/internal/interface/handler"
	"github.com/takumi-1234/OpenRAGLecture/internal/tests/mocks"
	"github.com/takumi-1234/OpenRAGLecture/internal/usecase/input"
	"github.com/takumi-1234/OpenRAGLecture/internal/usecase/output"
	appErrors "github.com/takumi-1234/OpenRAGLecture/pkg/errors"
)

func TestAuthHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAuthUsecase := new(mocks.MockAuthUsecase)
	authHandler := handler.NewAuthHandler(mockAuthUsecase)
	router := gin.New()
	router.POST("/register", authHandler.Register)

	t.Run("Success", func(t *testing.T) {
		registerInput := input.RegisterInput{Email: "new@example.com", Password: "password123"}
		mockAuthUsecase.On("Register", mock.Anything, registerInput).Return(nil).Once()

		body, _ := json.Marshal(registerInput)
		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		mockAuthUsecase.AssertExpectations(t)
	})

	t.Run("Failure_Conflict", func(t *testing.T) {
		registerInput := input.RegisterInput{Email: "exists@example.com", Password: "password123"}
		mockAuthUsecase.On("Register", mock.Anything, registerInput).Return(appErrors.ErrEmailAlreadyExists).Once()

		body, _ := json.Marshal(registerInput)
		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusConflict, rr.Code)
		mockAuthUsecase.AssertExpectations(t)
	})

	t.Run("Failure_BadRequest", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"email": "bad@example.com"})
		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAuthUsecase := new(mocks.MockAuthUsecase)
	authHandler := handler.NewAuthHandler(mockAuthUsecase)
	router := gin.New()
	router.POST("/login", authHandler.Login)

	t.Run("Success", func(t *testing.T) {
		loginInput := input.LoginInput{Email: "user@example.com", Password: "password123"}
		loginOutput := &output.LoginOutput{AccessToken: "access-token", RefreshToken: "refresh-token"}
		mockAuthUsecase.On("Login", mock.Anything, loginInput).Return(loginOutput, nil).Once()

		body, _ := json.Marshal(loginInput)
		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var respBody output.LoginOutput
		json.Unmarshal(rr.Body.Bytes(), &respBody)
		assert.Equal(t, loginOutput.AccessToken, respBody.AccessToken)
		mockAuthUsecase.AssertExpectations(t)
	})

	t.Run("Failure_Unauthorized", func(t *testing.T) {
		loginInput := input.LoginInput{Email: "user@example.com", Password: "wrongpassword"}
		mockAuthUsecase.On("Login", mock.Anything, loginInput).Return(nil, appErrors.ErrInvalidCredentials).Once()

		body, _ := json.Marshal(loginInput)
		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		mockAuthUsecase.AssertExpectations(t)
	})
}
