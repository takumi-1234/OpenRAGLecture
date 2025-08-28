// OpenRAGLecture/internal/interface/handler/file_handler.go
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/takumi-1234/OpenRAGLecture/internal/usecase/port"
)

type FileHandler struct {
	fileUsecase port.FileUsecase
}

func NewFileHandler(fileUsecase port.FileUsecase) *FileHandler {
	return &FileHandler{fileUsecase: fileUsecase}
}

func (h *FileHandler) Upload(c *gin.Context) {
	courseIDStr := c.PostForm("course_id")
	courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course_id"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File not provided"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer file.Close()

	// TODO: Add authorization check to ensure the user is an instructor for this course.

	doc, err := h.fileUsecase.Upload(c.Request.Context(), courseID, fileHeader.Filename, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "File uploaded successfully. Processing has started.",
		"document_id": doc.ID,
	})
}
