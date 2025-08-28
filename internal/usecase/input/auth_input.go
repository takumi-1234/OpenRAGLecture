// OpenRAGLecture/internal/usecase/input/auth_input.go
package input

// RegisterInput represents the data needed to register a new user.
type RegisterInput struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name"`
}

// LoginInput represents the data needed for a user to log in.
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
