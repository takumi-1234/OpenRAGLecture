// OpenRAGLecture/internal/interface/handler/middleware/auth_middleware.go
package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/takumi-1234/OpenRAGLecture/pkg/auth"
	appErrors "github.com/takumi-1234/OpenRAGLecture/pkg/errors"
)

// AuthMiddleware creates a Gin middleware for JWT authentication.
func AuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			return
		}

		tokenString := parts[1]
		claims, err := jwtManager.Verify(tokenString)
		if err != nil {
			log.Printf("AuthMiddleware: token verification failed: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": appErrors.ErrUnauthorized.Error()})
			return
		}

		// Set user ID in context for downstream handlers using Gin's context
		c.Set(auth.UserIDKey, claims.UserID)

		c.Next()
	}
}
