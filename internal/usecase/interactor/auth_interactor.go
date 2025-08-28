// OpenRAGLecture/internal/usecase/interactor/auth_interactor.go
package interactor

import (
	"context"
	"database/sql"
	"errors"

	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/repository"
	appErrors "github.com/takumi-1234/OpenRAGLecture/pkg/errors"

	"github.com/takumi-1234/OpenRAGLecture/internal/usecase/input"
	"github.com/takumi-1234/OpenRAGLecture/internal/usecase/output"
	"github.com/takumi-1234/OpenRAGLecture/internal/usecase/port"
	"github.com/takumi-1234/OpenRAGLecture/pkg/auth"
)

type authInteractor struct {
	userRepo   repository.UserRepository
	jwtManager *auth.JWTManager
}

// NewAuthInteractor creates a new instance of AuthUsecase.
func NewAuthInteractor(userRepo repository.UserRepository, jwtManager *auth.JWTManager) port.AuthUsecase {
	return &authInteractor{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// Register handles user registration.
func (i *authInteractor) Register(ctx context.Context, in input.RegisterInput) error {
	// Check if user with the same email already exists
	existingUser, err := i.userRepo.FindByEmail(ctx, in.Email)
	if err != nil && !errors.Is(err, appErrors.ErrUserNotFound) {
		return appErrors.ErrInternalServerError
	}
	if existingUser != nil {
		return appErrors.ErrEmailAlreadyExists
	}

	// Create a new user
	user := &model.User{
		Email: in.Email,
		Role:  model.RoleStudent, // Default role
		// ★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★
		// 修正点: DisplayNameを適切に設定する
		// ★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★
		DisplayName: sql.NullString{String: in.DisplayName, Valid: in.DisplayName != ""},
	}
	if err := user.SetPassword(in.Password); err != nil {
		return appErrors.ErrInternalServerError
	}

	if err := i.userRepo.Create(ctx, user); err != nil {
		return appErrors.ErrInternalServerError
	}

	return nil
}

// Login handles user login.
func (i *authInteractor) Login(ctx context.Context, in input.LoginInput) (*output.LoginOutput, error) {
	user, err := i.userRepo.FindByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, appErrors.ErrUserNotFound) {
			return nil, appErrors.ErrInvalidCredentials
		}
		return nil, appErrors.ErrInternalServerError
	}

	if !user.CheckPassword(in.Password) {
		return nil, appErrors.ErrInvalidCredentials
	}

	// Generate JWT tokens
	accessToken, refreshToken, err := i.jwtManager.GenerateTokens(user)
	if err != nil {
		return nil, appErrors.ErrInternalServerError
	}

	return &output.LoginOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
