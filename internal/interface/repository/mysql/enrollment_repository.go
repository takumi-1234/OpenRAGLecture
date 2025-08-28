// open-rag-lecture/internal/interface/repository/mysql/enrollment_repository.go
package mysql

import (
	"context"

	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/repository"
	"gorm.io/gorm"
)

type enrollmentRepository struct {
	db *gorm.DB
}

// NewEnrollmentRepository creates a new EnrollmentRepository implementation.
func NewEnrollmentRepository(db *gorm.DB) repository.EnrollmentRepository {
	return &enrollmentRepository{db: db}
}

func (r *enrollmentRepository) Create(ctx context.Context, enrollment *model.Enrollment) error {
	return r.db.WithContext(ctx).Create(enrollment).Error
}

func (r *enrollmentRepository) IsEnrolled(ctx context.Context, userID, courseID uint64) (bool, error) {
	var count int64
	// GORM's default behavior for Count on a model with DeletedAt will automatically add `deleted_at IS NULL`.
	// This correctly checks for active enrollments.
	err := r.db.WithContext(ctx).Model(&model.Enrollment{}).
		Where("user_id = ? AND course_id = ?", userID, courseID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
