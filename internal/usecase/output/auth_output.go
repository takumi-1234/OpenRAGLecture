// OpenRAGLecture/internal/usecase/output/auth_output.go
package output

// LoginOutput represents the data returned upon successful login.
type LoginOutput struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
