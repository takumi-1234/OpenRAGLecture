// open-rag-lecture/internal/domain/repository/enrollment_repository.go
package repository

import (
	"context"

	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
)

// EnrollmentRepository defines the interface for enrollment-related database operations.
type EnrollmentRepository interface {
	// Create creates a new enrollment record.
	Create(ctx context.Context, enrollment *model.Enrollment) error
	// IsEnrolled checks if a user is already enrolled in a course.
	// It should check for non-soft-deleted records.
	IsEnrolled(ctx context.Context, userID, courseID uint64) (bool, error)
}
