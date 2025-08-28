// OpenRAGLecture/internal/usecase/port/qa_port.go
package port

import (
	"context"
	"io"

	"github.com/takumi-1234/OpenRAGLecture/internal/usecase/input"
)

// QAUsecase defines the interface for the core question-answering logic.
type QAUsecase interface {
	Ask(ctx context.Context, in input.AskInput) (string, error)
	AskStream(ctx context.Context, in input.AskInput, writer io.Writer) error
}
