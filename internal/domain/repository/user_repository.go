// OpenRAGLecture/internal/domain/repository/user_repository.go
package repository

import (
	"context"

	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
)

type UserRepository interface {
	FindByID(ctx context.Context, id uint64) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
}
