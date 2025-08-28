// OpenRAGLecture/internal/interface/handler/auth_handler.go
package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/takumi-1234/OpenRAGLecture/internal/usecase/input"
	"github.com/takumi-1234/OpenRAGLecture/internal/usecase/port"
	appErrors "github.com/takumi-1234/OpenRAGLecture/pkg/errors"
)

type AuthHandler struct {
	authUsecase port.AuthUsecase
}

func NewAuthHandler(authUsecase port.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var in input.RegisterInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": appErrors.ErrBadRequest.Error()})
		return
	}

	err := h.authUsecase.Register(c.Request.Context(), in)
	if err != nil {
		if errors.Is(err, appErrors.ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": appErrors.ErrInternalServerError.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var in input.LoginInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": appErrors.ErrBadRequest.Error()})
		return
	}

	out, err := h.authUsecase.Login(c.Request.Context(), in)
	if err != nil {
		if errors.Is(err, appErrors.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": appErrors.ErrInternalServerError.Error()})
		return
	}

	c.JSON(http.StatusOK, out)
}
