// OpenRAGLecture/internal/domain/repository/cache_repository.go
package repository

import (
	"context"
	"time"
)

type CacheRepository interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
}
