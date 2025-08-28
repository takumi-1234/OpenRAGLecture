// open-rag-lecture/internal/usecase/interactor/file_interactor.go

package interactor

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/repository"
	"github.com/takumi-1234/OpenRAGLecture/internal/usecase/port"
	appErrors "github.com/takumi-1234/OpenRAGLecture/pkg/errors"
)

type fileInteractor struct {
	docRepo     repository.DocumentRepository
	fileStorage repository.FileStorage
	courseRepo  repository.CourseRepository // ★ 依存関係に CourseRepository を追加
}

// NewFileInteractor creates a new instance of FileUsecase.
// ★ courseRepo を引数に追加
func NewFileInteractor(
	docRepo repository.DocumentRepository,
	fileStorage repository.FileStorage,
	courseRepo repository.CourseRepository,
) port.FileUsecase {
	return &fileInteractor{
		docRepo:     docRepo,
		fileStorage: fileStorage,
		courseRepo:  courseRepo,
	}
}

func (i *fileInteractor) Upload(ctx context.Context, courseID uint64, fileName string, file io.Reader) (*model.Document, error) {
	// Read the file content into a buffer to use it multiple times
	var buf bytes.Buffer
	tee := io.TeeReader(file, &buf)

	// Calculate checksum
	hash := sha256.New()
	if _, err := io.Copy(hash, tee); err != nil {
		return nil, fmt.Errorf("%w: %v", appErrors.ErrFileProcessingFailed, err)
	}
	checksum := hex.EncodeToString(hash.Sum(nil))

	// ★★★ ここからが修正ロジック ★★★
	// 1. courseIDを使ってCourseの完全な情報をDBから取得する
	course, err := i.courseRepo.FindByID(ctx, courseID)
	if err != nil {
		// courseが存在しない場合 (ErrCourseNotFound) やDBエラーの場合
		return nil, err
	}
	// ★★★ 修正ロジックここまで ★★★

	// Save the file using the file storage interface
	filePath, err := i.fileStorage.Save(ctx, courseID, fileName, buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", appErrors.ErrFileUploadFailed, err)
	}

	// Create a new document record in the database
	doc := &model.Document{
		CourseID:   courseID,
		SemesterID: course.SemesterID, // ★ 取得したcourse情報からSemesterIDを動的に設定
		Title:      fileName,
		SourceURI:  filePath,
		DocType:    model.DocTypePDF, // Assuming PDF, can be detected from mime-type
		Checksum:   checksum,
		Version:    1,
	}

	if err := i.docRepo.Create(ctx, doc); err != nil {
		// If DB write fails, try to clean up the uploaded file
		_ = i.fileStorage.Delete(ctx, filePath)
		return nil, appErrors.ErrInternalServerError
	}

	// TODO: Trigger asynchronous batch processing (chunking, embedding) here.
	// For example, by sending a message to a queue.

	return doc, nil
}

func (i *fileInteractor) Download(ctx context.Context, documentID uint64) ([]byte, *model.Document, error) {
	// Get document metadata from the database
	doc, err := i.docRepo.FindByID(ctx, documentID)
	if err != nil {
		return nil, nil, err // Can be ErrNotFound or ErrInternalServerError
	}

	// Get the file content from storage
	data, err := i.fileStorage.Get(ctx, doc.SourceURI)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: failed to retrieve file from storage", appErrors.ErrInternalServerError)
	}

	return data, doc, nil
}
