// internal/interface/router/router.go
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/takumi-1234/OpenRAGLecture/internal/interface/handler"
	"github.com/takumi-1234/OpenRAGLecture/internal/interface/handler/middleware"
	"github.com/takumi-1234/OpenRAGLecture/pkg/auth"
	"github.com/takumi-1234/OpenRAGLecture/pkg/config"
)

func NewRouter(
	cfg config.ServerConfig,
	authHandler *handler.AuthHandler,
	qaHandler *handler.QAHandler,
	fileHandler *handler.FileHandler,
	courseHandler *handler.CourseHandler,
	// ★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★
	// 修正点: 引数にHealthHandlerを追加
	// ★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★
	healthHandler *handler.HealthHandler,
	jwtManager *auth.JWTManager,
) *gin.Engine {
	gin.SetMode(cfg.Mode)
	router := gin.Default()

	// ★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★
	// 修正点: 認証不要のヘルスチェックエンドポイントを追加
	// ★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★
	router.GET("/healthz", healthHandler.Check)

	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/register", authHandler.Register)
		authRoutes.POST("/login", authHandler.Login)
	}

	apiRoutes := router.Group("/api")
	apiRoutes.Use(middleware.AuthMiddleware(jwtManager))
	{
		qaRoutes := apiRoutes.Group("/qa")
		{
			qaRoutes.POST("/ask", qaHandler.Ask)
		}

		fileRoutes := apiRoutes.Group("/files")
		{
			fileRoutes.POST("/upload", fileHandler.Upload)
		}

		courseRoutes := apiRoutes.Group("/courses")
		{
			courseRoutes.POST("/:course_id/enrollments", courseHandler.Enroll)
		}
	}

	return router
}
