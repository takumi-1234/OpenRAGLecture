// OpenRAGLecture/pkg/errors/errors.go

package errors

import "errors"

var (
	// Generic errors
	ErrInternalServerError = errors.New("internal server error")
	ErrBadRequest          = errors.New("bad request")
	ErrNotFound            = errors.New("resource not found")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrConflict            = errors.New("resource conflict or duplicate entry")
	ErrValidation          = errors.New("validation failed")

	// Specific errors
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrEmailAlreadyExists   = errors.New("email already exists")
	ErrUserNotFound         = errors.New("user not found")
	ErrDocumentNotFound     = errors.New("document not found")
	ErrCourseNotFound       = errors.New("course not found")
	ErrNotEnrolled          = errors.New("user not enrolled in this course")
	ErrFileUploadFailed     = errors.New("file upload failed")
	ErrFileProcessingFailed = errors.New("file processing failed")
)
