// OpenRAGLecture/internal/interface/repository/storage/local_storage.go
package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/repository"
	"github.com/takumi-1234/OpenRAGLecture/pkg/config"
)

type localStorage struct {
	basePath string
}

// NewLocalStorage creates a new FileStorage implementation for the local filesystem.
func NewLocalStorage(cfg config.LocalStorageConfig) (repository.FileStorage, error) {
	path := cfg.Path
	// The complex path mapping for tests is removed for simplification.
	// We will rely on consistent path management within the application.
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create local storage directory: %w", err)
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for local storage: %w", err)
	}
	return &localStorage{basePath: absPath}, nil
}

func (s *localStorage) Save(_ context.Context, courseID uint64, fileName string, data []byte) (string, error) {
	// Create a unique file name to avoid collisions
	uniqueFileName := fmt.Sprintf("%s-%s", uuid.New().String(), filepath.Base(fileName))

	// Organize files by course, relative to the base path
	relativeCoursePath := filepath.Join(fmt.Sprintf("%d", courseID))
	fullCoursePath := filepath.Join(s.basePath, relativeCoursePath)

	if err := os.MkdirAll(fullCoursePath, 0755); err != nil {
		return "", err
	}

	relativePath := filepath.Join(relativeCoursePath, uniqueFileName)
	fullPath := filepath.Join(s.basePath, relativePath)

	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return "", err
	}

	// Return the path relative to the base storage directory
	return relativePath, nil
}

func (s *localStorage) Get(_ context.Context, path string) ([]byte, error) {
	// Prevent path traversal attacks.
	// The path should be a relative path within the basePath.
	cleanRelativePath := filepath.Clean(path)
	if strings.HasPrefix(cleanRelativePath, "..") {
		return nil, os.ErrNotExist
	}

	fullPath := filepath.Join(s.basePath, cleanRelativePath)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		// To be safe, return a generic error to avoid leaking path information.
		if os.IsNotExist(err) {
			return nil, os.ErrNotExist
		}
		return nil, err
	}
	return data, nil
}

func (s *localStorage) Delete(_ context.Context, path string) error {
	// Prevent path traversal attacks
	cleanRelativePath := filepath.Clean(path)
	if strings.HasPrefix(cleanRelativePath, "..") {
		return os.ErrNotExist
	}

	fullPath := filepath.Join(s.basePath, cleanRelativePath)

	err := os.Remove(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return os.ErrNotExist
		}
		return err
	}
	return nil
}