// internal/tests/usecase/qa_interactor_test.go
package usecase_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/repository"
	"github.com/takumi-1234/OpenRAGLecture/internal/tests/mocks"
	"github.com/takumi-1234/OpenRAGLecture/internal/usecase/input"
	"github.com/takumi-1234/OpenRAGLecture/internal/usecase/interactor"
)

func TestQAInteractor_Ask(t *testing.T) {
	ctx := context.Background()
	mockDocRepo := new(mocks.MockDocumentRepository)
	mockVectorRepo := new(mocks.MockVectorRepository)
	mockEmbeddingRepo := new(mocks.MockEmbeddingRepository)
	mockLLMRepo := new(mocks.MockLLMRepository)

	qaInteractor := interactor.NewQAInteractor(
		mockDocRepo,
		mockVectorRepo,
		mockEmbeddingRepo,
		mockLLMRepo,
	)

	askInput := input.AskInput{
		UserID:   1,
		CourseID: 101,
		Query:    "What is RAG?",
	}

	queryVector := []float32{0.1, 0.2, 0.3}

	vectorResults := []model.RetrievedChunk{
		{Chunk: model.Chunk{Base: model.Base{ID: 1}, Text: "RAG stands for Retrieval-Augmented Generation."}, Score: 0.9},
	}
	bm25Results := []model.RetrievedChunk{
		{Chunk: model.Chunk{Base: model.Base{ID: 2}, Text: "This document mentions RAG."}, Score: 0.8},
	}

	t.Run("Success_HappyPath", func(t *testing.T) {
		// Arrange

		// ★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★
		// 修正点: contextの期待値を mock.Anything に変更
		// これにより、errgroupが生成する *context.cancelCtx 型にもマッチするようになります。
		// ★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★
		mockEmbeddingRepo.On("CreateEmbeddings", mock.Anything, []string{askInput.Query}, "RETRIEVAL_QUERY").Return([][]float32{queryVector}, nil).Once()
		mockVectorRepo.On("Search", mock.Anything, queryVector, askInput.CourseID, 5).Return(vectorResults, nil).Once()
		mockDocRepo.On("FullTextSearch", mock.Anything, askInput.Query, askInput.CourseID, 5).Return(bm25Results, nil).Once()

		expectedLLMAnswer := "Retrieval-Augmented Generation (RAG) is a technique..."
		mockLLMRepo.On("GenerateContent", mock.Anything, mock.AnythingOfType("repository.GenerateContentParams")).
			Return(expectedLLMAnswer, nil).
			Run(func(args mock.Arguments) {
				params := args.Get(1).(repository.GenerateContentParams)
				assert.Contains(t, params.SystemPrompt, "RAG stands for")
				assert.Equal(t, askInput.Query, params.UserPrompt)
			}).Once()

		// Act
		answer, err := qaInteractor.Ask(ctx, askInput)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedLLMAnswer, answer)
		mockEmbeddingRepo.AssertExpectations(t)
		mockVectorRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
		mockLLMRepo.AssertExpectations(t)
	})

	t.Run("Success_NoResultsFound", func(t *testing.T) {
		// Arrange
		mockEmbeddingRepo.On("CreateEmbeddings", mock.Anything, []string{askInput.Query}, "RETRIEVAL_QUERY").Return([][]float32{queryVector}, nil).Once()
		mockVectorRepo.On("Search", mock.Anything, queryVector, askInput.CourseID, 5).Return([]model.RetrievedChunk{}, nil).Once()
		mockDocRepo.On("FullTextSearch", mock.Anything, askInput.Query, askInput.CourseID, 5).Return([]model.RetrievedChunk{}, nil).Once()

		// Act
		answer, err := qaInteractor.Ask(ctx, askInput)

		// Assert
		assert.NoError(t, err)
		assert.True(t, strings.HasPrefix(answer, "I could not find any relevant information"))
		mockLLMRepo.AssertNotCalled(t, "GenerateContent")
	})

	t.Run("Failure_EmbeddingGenerationFails", func(t *testing.T) {
		// Arrange
		mockEmbeddingRepo.On("CreateEmbeddings", mock.Anything, []string{askInput.Query}, "RETRIEVAL_QUERY").Return(nil, fmt.Errorf("API error")).Once()

		// Act
		_, err := qaInteractor.Ask(ctx, askInput)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create query embedding")
		mockVectorRepo.AssertNotCalled(t, "Search")
		mockDocRepo.AssertNotCalled(t, "FullTextSearch")
	})
}
