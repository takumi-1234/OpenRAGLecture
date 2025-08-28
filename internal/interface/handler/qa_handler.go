// open-rag-lecture/internal/interface/handler/qa_handler.go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/repository"
	"github.com/takumi-1234/OpenRAGLecture/internal/usecase/input"
	"github.com/takumi-1234/OpenRAGLecture/internal/usecase/port"
	"github.com/takumi-1234/OpenRAGLecture/pkg/auth"
	appErrors "github.com/takumi-1234/OpenRAGLecture/pkg/errors"
)

type QAHandler struct {
	qaUsecase      port.QAUsecase
	enrollmentRepo repository.EnrollmentRepository // ★ 修正: CourseRepoからEnrollmentRepoへ
}

// ★ 修正: NewQAHandler の引数を変更
func NewQAHandler(qaUsecase port.QAUsecase, enrollmentRepo repository.EnrollmentRepository) *QAHandler {
	return &QAHandler{
		qaUsecase:      qaUsecase,
		enrollmentRepo: enrollmentRepo,
	}
}

func (h *QAHandler) Ask(c *gin.Context) {
	var in input.AskInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": appErrors.ErrBadRequest.Error()})
		return
	}

	userID, ok := auth.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found in context"})
		return
	}
	in.UserID = userID

	// ★ 修正: enrollmentRepo.IsEnrolled を呼び出す
	isEnrolled, err := h.enrollmentRepo.IsEnrolled(c.Request.Context(), in.UserID, in.CourseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check enrollment"})
		return
	}
	if !isEnrolled {
		c.JSON(http.StatusForbidden, gin.H{"error": appErrors.ErrNotEnrolled.Error()})
		return
	}

	response, err := h.qaUsecase.Ask(c.Request.Context(), in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get answer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"answer": response})
}
