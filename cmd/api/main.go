// cmd/api/main.go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/takumi-1234/OpenRAGLecture/internal/interface/handler"
	"github.com/takumi-1234/OpenRAGLecture/internal/interface/repository/google"
	"github.com/takumi-1234/OpenRAGLecture/internal/interface/repository/mysql"
	"github.com/takumi-1234/OpenRAGLecture/internal/interface/repository/qdrant"
	"github.com/takumi-1234/OpenRAGLecture/internal/interface/repository/storage"
	"github.com/takumi-1234/OpenRAGLecture/internal/interface/router"
	"github.com/takumi-1234/OpenRAGLecture/internal/usecase/interactor"
	"github.com/takumi-1234/OpenRAGLecture/pkg/auth"
	"github.com/takumi-1234/OpenRAGLecture/pkg/config"
	"github.com/takumi-1234/OpenRAGLecture/pkg/logger"
)

// ★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★
// 修正点: アプリケーション起動時の自動マイグレーションを削除します。
// スキーマ管理は `golang-migrate` と `Makefile` を通じて明示的に行います。
// これにより、テストと本番環境でのスキーマの一貫性が保証され、
// 意図しないスキーマ変更やテスト時の競合を防ぎます。
// func runAutoMigration(db *gorm.DB) { ... } は削除
// ★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★

func main() {
	cfg, err := config.LoadConfig("./configs")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	appLogger, err := logger.NewLogger(cfg.Logging.Level, cfg.Logging.Encoding)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer appLogger.Sync()

	db, err := mysql.NewGORMClient(cfg.Database.MySQL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// ★★★ runAutoMigration(db) の呼び出しを削除 ★★★

	qdrantRepo, err := qdrant.NewQdrantRepository(cfg.VectorDB.Qdrant)
	if err != nil {
		log.Fatalf("Failed to connect to Qdrant: %v", err)
	}

	googleEmbeddingRepo, err := google.NewGoogleEmbeddingRepository(cfg.Google)
	if err != nil {
		log.Fatalf("Failed to create Google Embedding repo: %v", err)
	}

	googleLLMRepo, err := google.NewGoogleLLMRepository(cfg.Google)
	if err != nil {
		log.Fatalf("Failed to create Google LLM repo: %v", err)
	}

	fileStorage, err := storage.NewLocalStorage(cfg.Storage.Local)
	if err != nil {
		log.Fatalf("Failed to init file storage: %v", err)
	}

	// Repositories
	userRepo := mysql.NewUserRepository(db)
	docRepo := mysql.NewDocumentRepository(db)
	courseRepo := mysql.NewCourseRepository(db)
	feedbackRepo := mysql.NewFeedbackRepository(db)
	enrollmentRepo := mysql.NewEnrollmentRepository(db)

	// JWT Manager
	accessTokenDuration := time.Duration(cfg.Auth.JWT.AccessTokenExpiryHours) * time.Hour
	refreshTokenDuration := time.Duration(cfg.Auth.JWT.RefreshTokenExpiryDays) * 24 * time.Hour
	jwtManager := auth.NewJWTManager(
		cfg.Auth.JWT.SecretKey,
		accessTokenDuration,
		refreshTokenDuration,
	)

	// Usecases
	authUsecase := interactor.NewAuthInteractor(userRepo, jwtManager)
	qaUsecase := interactor.NewQAInteractor(docRepo, qdrantRepo, googleEmbeddingRepo, googleLLMRepo)
	fileUsecase := interactor.NewFileInteractor(docRepo, fileStorage, courseRepo)
	courseUsecase := interactor.NewCourseInteractor(courseRepo, enrollmentRepo)
	_ = interactor.NewFeedbackInteractor(feedbackRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authUsecase)
	qaHandler := handler.NewQAHandler(qaUsecase, enrollmentRepo)
	fileHandler := handler.NewFileHandler(fileUsecase)
	courseHandler := handler.NewCourseHandler(courseUsecase)
	healthHandler := handler.NewHealthHandler(db)

	// Router
	appRouter := router.NewRouter(cfg.Server, authHandler, qaHandler, fileHandler, courseHandler, healthHandler, jwtManager)

	serverAddr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Starting server on %s\n", serverAddr)
	if err := appRouter.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
