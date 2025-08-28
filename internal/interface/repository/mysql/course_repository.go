// open-rag-lecture/internal/interface/repository/mysql/course_repository.go
package mysql

import (
	"context"
	"errors"

	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/repository"
	appErrors "github.com/takumi-1234/OpenRAGLecture/pkg/errors"
	"gorm.io/gorm"
)

type courseRepository struct {
	db *gorm.DB
}

// NewCourseRepository creates a new CourseRepository implementation.
func NewCourseRepository(db *gorm.DB) repository.CourseRepository {
	return &courseRepository{db: db}
}

func (r *courseRepository) FindByID(ctx context.Context, id uint64) (*model.Course, error) {
	var course model.Course
	if err := r.db.WithContext(ctx).First(&course, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrCourseNotFound
		}
		return nil, err
	}
	return &course, nil
}

// CheckEnrollment は enrollmentRepository に移管されました
