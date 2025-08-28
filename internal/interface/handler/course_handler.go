// open-rag-lecture/internal/interface/handler/course_handler.go
package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/takumi-1234/OpenRAGLecture/internal/usecase/port"
	"github.com/takumi-1234/OpenRAGLecture/pkg/auth"
	appErrors "github.com/takumi-1234/OpenRAGLecture/pkg/errors"
)

type CourseHandler struct {
	courseUsecase port.CourseUsecase
}

func NewCourseHandler(courseUsecase port.CourseUsecase) *CourseHandler {
	return &CourseHandler{courseUsecase: courseUsecase}
}

// Enroll handles the request to enroll a user in a course.
func (h *CourseHandler) Enroll(c *gin.Context) {
	// 1. Get course_id from URL parameter.
	courseIDStr := c.Param("course_id")
	courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course_id format"})
		return
	}

	// 2. Get user_id from JWT claims in the context.
	userID, ok := auth.GetUserIDFromContext(c)
	if !ok {
		// This should technically be caught by the middleware, but as a safeguard:
		c.JSON(http.StatusUnauthorized, gin.H{"error": appErrors.ErrUnauthorized.Error()})
		return
	}

	// 3. Call the usecase.
	err = h.courseUsecase.EnrollUser(c.Request.Context(), userID, courseID)
	if err != nil {
		switch {
		case errors.Is(err, appErrors.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Course not found"})
		case errors.Is(err, appErrors.ErrConflict):
			c.JSON(http.StatusConflict, gin.H{"error": "User is already enrolled in this course"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": appErrors.ErrInternalServerError.Error()})
		}
		return
	}

	// 4. Return success response.
	c.Status(http.StatusCreated)
}
