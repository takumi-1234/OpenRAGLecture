// open-rag-lecture/internal/usecase/port/course_port.go
package port

import (
	"context"
)

// CourseUsecase defines the interface for course-related business logic, including enrollment.
type CourseUsecase interface {
	// EnrollUser enrolls the given user to the specified course.
	EnrollUser(ctx context.Context, userID, courseID uint64) error
}
