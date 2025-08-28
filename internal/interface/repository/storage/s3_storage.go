// OpenRAGLecture/internal/interface/repository/storage/s3_storage.go
package storage

import (
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/repository"
	"github.com/takumi-1234/OpenRAGLecture/pkg/config"
)

// NewS3Storage creates a new FileStorage implementation for S3.
func NewS3Storage(cfg config.S3StorageConfig) (repository.FileStorage, error) {
	// TODO: Implement S3 storage
	return nil, nil
}
