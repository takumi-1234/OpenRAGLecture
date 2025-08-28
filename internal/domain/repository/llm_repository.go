// OpenRAGLecture/internal/domain/repository/llm_repository.go
package repository

import (
	"context"
	"io"

	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
)

type GenerateContentParams struct {
	SystemPrompt  string
	UserPrompt    string
	ContextChunks []model.RetrievedChunk
}

type LLMRepository interface {
	// GenerateContent generates a response from the LLM based on the provided context.
	GenerateContent(ctx context.Context, params GenerateContentParams) (string, error)
	// GenerateContentStream generates a response as a stream.
	GenerateContentStream(ctx context.Context, params GenerateContentParams, writer io.Writer) error
}
