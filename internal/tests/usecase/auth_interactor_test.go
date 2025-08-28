// internal/tests/usecase/auth_interactor_test.go
package usecase_test // ★★★★★★★【最重要】 "package usecase" から "package usecase_test" に修正 ★★★★★★★

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
	"github.com/takumi-1234/OpenRAGLecture/internal/tests/mocks"
	"github.com/takumi-1234/OpenRAGLecture/internal/usecase/input"
	"github.com/takumi-1234/OpenRAGLecture/internal/usecase/interactor"
	"github.com/takumi-1234/OpenRAGLecture/pkg/auth"
	appErrors "github.com/takumi-1234/OpenRAGLecture/pkg/errors"
)

func TestAuthInteractor_Register(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := new(mocks.MockUserRepository)
	jwtManager := auth.NewJWTManager("test-secret", time.Hour, time.Hour*24)
	authInteractor := interactor.NewAuthInteractor(mockUserRepo, jwtManager)

	registerInput := input.RegisterInput{
		Email:    "test@example.com",
		Password: "password123",
	}

	t.Run("Success", func(t *testing.T) {
		mockUserRepo.On("FindByEmail", ctx, registerInput.Email).Return(nil, appErrors.ErrUserNotFound).Once()
		mockUserRepo.On("Create", ctx, mock.AnythingOfType("*model.User")).Return(nil).Once()

		err := authInteractor.Register(ctx, registerInput)

		assert.NoError(t, err)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Failure_EmailAlreadyExists", func(t *testing.T) {
		existingUser := &model.User{Email: registerInput.Email}
		mockUserRepo.On("FindByEmail", ctx, registerInput.Email).Return(existingUser, nil).Once()

		err := authInteractor.Register(ctx, registerInput)

		assert.ErrorIs(t, err, appErrors.ErrEmailAlreadyExists)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestAuthInteractor_Login(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := new(mocks.MockUserRepository)
	jwtManager := auth.NewJWTManager("test-secret", time.Hour, time.Hour*24)
	authInteractor := interactor.NewAuthInteractor(mockUserRepo, jwtManager)

	loginInput := input.LoginInput{
		Email:    "test@example.com",
		Password: "password123",
	}

	user := &model.User{
		Base:  model.Base{ID: 1},
		Email: loginInput.Email,
		Role:  model.RoleStudent,
	}
	_ = user.SetPassword(loginInput.Password)

	t.Run("Success", func(t *testing.T) {
		mockUserRepo.On("FindByEmail", ctx, loginInput.Email).Return(user, nil).Once()

		output, err := authInteractor.Login(ctx, loginInput)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.NotEmpty(t, output.AccessToken)
		assert.NotEmpty(t, output.RefreshToken)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Failure_UserNotFound", func(t *testing.T) {
		mockUserRepo.On("FindByEmail", ctx, loginInput.Email).Return(nil, appErrors.ErrUserNotFound).Once()

		_, err := authInteractor.Login(ctx, loginInput)

		assert.ErrorIs(t, err, appErrors.ErrInvalidCredentials)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Failure_InvalidPassword", func(t *testing.T) {
		userWithWrongPass := &model.User{
			Email:        loginInput.Email,
			PasswordHash: sql.NullString{String: "wronghash", Valid: true},
		}
		mockUserRepo.On("FindByEmail", ctx, loginInput.Email).Return(userWithWrongPass, nil).Once()

		_, err := authInteractor.Login(ctx, loginInput)

		assert.ErrorIs(t, err, appErrors.ErrInvalidCredentials)
		mockUserRepo.AssertExpectations(t)
	})
}
