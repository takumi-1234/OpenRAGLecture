// OpenRAGLecture/internal/usecase/port/file_port.go
package port

import (
	"context"
	"io"

	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
)

// FileUsecase defines the interface for file management logic.
type FileUsecase interface {
	Upload(ctx context.Context, courseID uint64, fileName string, file io.Reader) (*model.Document, error)
	Download(ctx context.Context, documentID uint64) ([]byte, *model.Document, error)
}
