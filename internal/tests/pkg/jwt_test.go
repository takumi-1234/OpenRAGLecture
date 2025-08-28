// internal/tests/pkg/jwt_test.go
package pkg_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
	"github.com/takumi-1234/OpenRAGLecture/pkg/auth"
)

func TestJWTManager(t *testing.T) {
	secretKey := "test-secret-key-that-is-sufficiently-long"
	tokenDuration := 15 * time.Minute
	refreshDuration := 24 * time.Hour
	jwtManager := auth.NewJWTManager(secretKey, tokenDuration, refreshDuration)

	user := &model.User{
		Base: model.Base{
			ID: 1,
		},
		Email: "test@example.com",
		Role:  model.RoleStudent,
		PasswordHash: sql.NullString{
			String: "hashedpassword",
			Valid:  true,
		},
	}

	t.Run("GenerateAndVerifyAccessToken_Success", func(t *testing.T) {
		accessToken, _, err := jwtManager.GenerateTokens(user)
		require.NoError(t, err)
		require.NotEmpty(t, accessToken)

		claims, err := jwtManager.Verify(accessToken)
		require.NoError(t, err)
		require.NotNil(t, claims)

		assert.Equal(t, user.ID, claims.UserID)
		assert.Equal(t, user.Email, claims.Email)
		assert.Equal(t, user.Role, claims.Role)
		assert.WithinDuration(t, time.Now().Add(tokenDuration), claims.ExpiresAt.Time, 5*time.Second)
	})

	t.Run("VerifyToken_Expired", func(t *testing.T) {
		shortLivedManager := auth.NewJWTManager(secretKey, -1*time.Minute, refreshDuration)
		accessToken, _, err := shortLivedManager.GenerateTokens(user)
		require.NoError(t, err)

		_, err = jwtManager.Verify(accessToken)
		assert.Error(t, err, "token is expired")
	})

	t.Run("VerifyToken_InvalidSignature", func(t *testing.T) {
		accessToken, _, err := jwtManager.GenerateTokens(user)
		require.NoError(t, err)

		anotherManager := auth.NewJWTManager("another-secret-key", tokenDuration, refreshDuration)
		_, err = anotherManager.Verify(accessToken)
		assert.Error(t, err, "signature is invalid")
	})
}
