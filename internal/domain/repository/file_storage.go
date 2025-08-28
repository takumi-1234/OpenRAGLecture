// OpenRAGLecture/internal/domain/repository/file_storage.go
package repository

import "context"

type FileStorage interface {
	Save(ctx context.Context, courseID uint64, fileName string, data []byte) (path string, err error)
	Get(ctx context.Context, path string) ([]byte, error)
	Delete(ctx context.Context, path string) error
}
