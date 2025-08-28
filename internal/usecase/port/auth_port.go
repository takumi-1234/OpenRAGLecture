// OpenRAGLecture/internal/usecase/port/auth_port.go
package port

import (
	"context"

	"github.com/takumi-1234/OpenRAGLecture/internal/usecase/input"
	"github.com/takumi-1234/OpenRAGLecture/internal/usecase/output"
)

// AuthUsecase defines the interface for authentication-related business logic.
type AuthUsecase interface {
	Register(ctx context.Context, in input.RegisterInput) error
	Login(ctx context.Context, in input.LoginInput) (*output.LoginOutput, error)
}
