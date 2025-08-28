// internal/tests/handler/helpers_test.go
package handler_test

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/takumi-1234/OpenRAGLecture/pkg/auth"
)

// authMiddlewareMock is a mock middleware to set user ID in context
func authMiddlewareMock(userID uint64) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), auth.UserIDKey, userID)
		c.Request = c.Request.WithContext(ctx)
		c.Set(auth.UserIDKey, userID) // Also set in Gin's context
		c.Next()
	}
}
