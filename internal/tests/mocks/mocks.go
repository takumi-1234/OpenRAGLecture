// internal/tests/mocks/mocks.go
package mocks

import (
	"context"
	"io"

	"github.com/stretchr/testify/mock"
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/repository"
	"github.com/takumi-1234/OpenRAGLecture/internal/usecase/input"
	"github.com/takumi-1234/OpenRAGLecture/internal/usecase/output"
)

// ==============================================================================
// REPOSITORY MOCKS
// ==============================================================================

// MockUserRepository is a mock of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uint64) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// MockCourseRepository is a mock of CourseRepository
type MockCourseRepository struct {
	mock.Mock
}

func (m *MockCourseRepository) FindByID(ctx context.Context, id uint64) (*model.Course, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Course), args.Error(1)
}

// MockEnrollmentRepository is a mock of EnrollmentRepository
type MockEnrollmentRepository struct {
	mock.Mock
}

func (m *MockEnrollmentRepository) Create(ctx context.Context, enrollment *model.Enrollment) error {
	args := m.Called(ctx, enrollment)
	return args.Error(0)
}

func (m *MockEnrollmentRepository) IsEnrolled(ctx context.Context, userID, courseID uint64) (bool, error) {
	args := m.Called(ctx, userID, courseID)
	return args.Bool(0), args.Error(1)
}

// MockDocumentRepository is a mock of DocumentRepository
type MockDocumentRepository struct {
	mock.Mock
}

func (m *MockDocumentRepository) FindByID(ctx context.Context, id uint64) (*model.Document, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Document), args.Error(1)
}

func (m *MockDocumentRepository) Create(ctx context.Context, doc *model.Document) error {
	args := m.Called(ctx, doc)
	return args.Error(0)
}

func (m *MockDocumentRepository) FullTextSearch(ctx context.Context, query string, courseID uint64, limit int) ([]model.RetrievedChunk, error) {
	args := m.Called(ctx, query, courseID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.RetrievedChunk), args.Error(1)
}

// MockEmbeddingRepository is a mock of EmbeddingRepository
type MockEmbeddingRepository struct {
	mock.Mock
}

func (m *MockEmbeddingRepository) CreateEmbeddings(ctx context.Context, texts []string, taskType string) ([][]float32, error) {
	args := m.Called(ctx, texts, taskType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([][]float32), args.Error(1)
}

// MockVectorRepository is a mock of VectorRepository
type MockVectorRepository struct {
	mock.Mock
}

func (m *MockVectorRepository) Upsert(ctx context.Context, chunks []*model.Chunk, vectors [][]float32) error {
	args := m.Called(ctx, chunks, vectors)
	return args.Error(0)
}

func (m *MockVectorRepository) Search(ctx context.Context, queryVector []float32, courseID uint64, limit int) ([]model.RetrievedChunk, error) {
	args := m.Called(ctx, queryVector, courseID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.RetrievedChunk), args.Error(1)
}

func (m *MockVectorRepository) RecreateCollection(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockVectorRepository) EnsureCollectionExists(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockLLMRepository is a mock of LLMRepository
type MockLLMRepository struct {
	mock.Mock
}

func (m *MockLLMRepository) GenerateContent(ctx context.Context, params repository.GenerateContentParams) (string, error) {
	args := m.Called(ctx, params)
	return args.String(0), args.Error(1)
}

func (m *MockLLMRepository) GenerateContentStream(ctx context.Context, params repository.GenerateContentParams, writer io.Writer) error {
	args := m.Called(ctx, params, writer)
	return args.Error(0)
}

// MockFileStorage is a mock of FileStorage
type MockFileStorage struct {
	mock.Mock
}

func (m *MockFileStorage) Save(ctx context.Context, courseID uint64, fileName string, data []byte) (string, error) {
	args := m.Called(ctx, courseID, fileName, data)
	return args.String(0), args.Error(1)
}

func (m *MockFileStorage) Get(ctx context.Context, path string) ([]byte, error) {
	args := m.Called(ctx, path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockFileStorage) Delete(ctx context.Context, path string) error {
	args := m.Called(ctx, path)
	return args.Error(0)
}

// ==============================================================================
// USECASE MOCKS
// ==============================================================================

// MockAuthUsecase is a mock of AuthUsecase
type MockAuthUsecase struct {
	mock.Mock
}

func (m *MockAuthUsecase) Register(ctx context.Context, in input.RegisterInput) error {
	args := m.Called(ctx, in)
	return args.Error(0)
}

func (m *MockAuthUsecase) Login(ctx context.Context, in input.LoginInput) (*output.LoginOutput, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*output.LoginOutput), args.Error(1)
}

// MockCourseUsecase is a mock of CourseUsecase
type MockCourseUsecase struct {
	mock.Mock
}

func (m *MockCourseUsecase) EnrollUser(ctx context.Context, userID, courseID uint64) error {
	args := m.Called(ctx, userID, courseID)
	return args.Error(0)
}

// MockFileUsecase is a mock of FileUsecase
type MockFileUsecase struct {
	mock.Mock
}

func (m *MockFileUsecase) Upload(ctx context.Context, courseID uint64, fileName string, file io.Reader) (*model.Document, error) {
	args := m.Called(ctx, courseID, fileName, file)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Document), args.Error(1)
}

func (m *MockFileUsecase) Download(ctx context.Context, documentID uint64) ([]byte, *model.Document, error) {
	args := m.Called(ctx, documentID)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]byte), args.Get(1).(*model.Document), args.Error(2)
}

// MockQAUsecase is a mock of QAUsecase
type MockQAUsecase struct {
	mock.Mock
}

func (m *MockQAUsecase) Ask(ctx context.Context, in input.AskInput) (string, error) {
	args := m.Called(ctx, in)
	return args.String(0), args.Error(1)
}

func (m *MockQAUsecase) AskStream(ctx context.Context, in input.AskInput, writer io.Writer) error {
	args := m.Called(ctx, in, writer)
	return args.Error(0)
}
