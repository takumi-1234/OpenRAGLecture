// open-rag-lecture/internal/domain/repository/course_repository.go
package repository

import (
	"context"

	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
)

type CourseRepository interface {
	FindByID(ctx context.Context, id uint64) (*model.Course, error)
	// CheckEnrollment は EnrollmentRepository に移管されました
}
