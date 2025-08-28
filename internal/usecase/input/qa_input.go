// OpenRAGLecture/internal/usecase/input/qa_input.go
package input

// AskInput represents the data for a user's question.
type AskInput struct {
	UserID    uint64 `json:"-"` // From JWT, not from request body
	CourseID  uint64 `json:"course_id" binding:"required"`
	Query     string `json:"query" binding:"required"`
	SessionID string `json:"session_id"` // For conversation history (optional)
}
