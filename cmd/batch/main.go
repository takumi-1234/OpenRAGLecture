// open-rag-lecture/cmd/batch/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/takumi-1234/OpenRAGLecture/internal/batch" // ★★★ 新しいパッケージをインポート
	"github.com/takumi-1234/OpenRAGLecture/pkg/config"
	"github.com/takumi-1234/OpenRAGLecture/pkg/logger"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.LoadConfig("./configs")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 2. Initialize Logger
	appLogger, err := logger.NewLogger(cfg.Logging.Level, "console")
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer appLogger.Sync()

	// 3. Get task name from command-line arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/batch/*.go <task_name>")
		fmt.Println("Available tasks: sync-documents")
		os.Exit(1)
	}
	taskName := os.Args[1]

	// 4. Build and run the task
	// ★★★ 新しいパッケージの NewTaskRunner を呼び出す
	runner, err := batch.NewTaskRunner(taskName, cfg)
	if err != nil {
		log.Fatalf("Failed to create task runner: %v", err)
	}

	if err := runner.Run(context.Background()); err != nil {
		log.Fatalf("Task '%s' failed: %v", taskName, err)
	}

	log.Printf("Task '%s' completed successfully.", taskName)
}