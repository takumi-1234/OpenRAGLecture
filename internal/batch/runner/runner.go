// open-rag-lecture/internal/batch/runner.go
package batch

import (
	"context"
	"fmt"

	"github.com/takumi-1234/OpenRAGLecture/internal/batch/processor"
	"github.com/takumi-1234/OpenRAGLecture/internal/batch/task"
	"github.com/takumi-1234/OpenRAGLecture/internal/interface/repository/google"
	"github.com/takumi-1234/OpenRAGLecture/internal/interface/repository/mysql"
	"github.com/takumi-1234/OpenRAGLecture/internal/interface/repository/qdrant"
	"github.com/takumi-1234/OpenRAGLecture/internal/interface/repository/storage"
	"github.com/takumi-1234/OpenRAGLecture/pkg/config"
)

// TaskRunner defines an interface for a runnable batch task.
type TaskRunner interface {
	Run(ctx context.Context) error
}

// NewTaskRunner builds and returns a specific task runner based on the task name.
func NewTaskRunner(taskName string, cfg config.Config) (TaskRunner, error) {
	// Initialize dependencies
	db, err := mysql.NewGORMClient(cfg.Database.MySQL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	qdrantRepo, err := qdrant.NewQdrantRepository(cfg.VectorDB.Qdrant)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Qdrant: %w", err)
	}

	googleEmbeddingRepo, err := google.NewGoogleEmbeddingRepository(cfg.Google)
	if err != nil {
		return nil, fmt.Errorf("failed to create Google Embedding repo: %w", err)
	}

	fileStorage, err := storage.NewLocalStorage(cfg.Storage.Local)
	if err != nil {
		return nil, fmt.Errorf("failed to init file storage: %w", err)
	}

	// ★★★ 修正点: 設定ファイルからモデル名を渡す ★★★
	chunkProc := processor.NewPDFChunkProcessor(cfg.Google.EmbeddingModel)

	// Return the requested task
	switch taskName {
	case "sync-documents":
		return task.NewSyncTask(db, fileStorage, chunkProc, googleEmbeddingRepo, qdrantRepo), nil
	default:
		return nil, fmt.Errorf("unknown task: %s", taskName)
	}
}
