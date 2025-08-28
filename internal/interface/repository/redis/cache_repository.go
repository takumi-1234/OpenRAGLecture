// OpenRAGLecture/internal/interface/repository/redis/cache_repository.go
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/repository"
	"github.com/takumi-1234/OpenRAGLecture/pkg/config"
)

type redisRepository struct {
	client *redis.Client
}

// NewRedisRepository creates a new CacheRepository implementation for Redis.
func NewRedisRepository(cfg config.RedisConfig) (repository.CacheRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Ping the server to check the connection
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &redisRepository{client: client}, nil
}

func (r *redisRepository) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // Not found is not an error
	}
	return val, err
}

func (r *redisRepository) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *redisRepository) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}
