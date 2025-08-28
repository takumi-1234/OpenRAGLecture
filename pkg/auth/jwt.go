// open-rag-lecture/pkg/auth/jwt.go

package auth

import (
	"fmt"
	"log" // ★★★ log パッケージをインポート ★★★
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/takumi-1234/OpenRAGLecture/internal/domain/model"
)

const UserIDKey = "userID"

// JWTManager manages JWT token generation and validation.
type JWTManager struct {
	secretKey       string
	tokenDuration   time.Duration
	refreshDuration time.Duration
}

// UserClaims is a custom claims structure that includes user details.
type UserClaims struct {
	UserID uint64     `json:"user_id"`
	Email  string     `json:"email"`
	Role   model.Role `json:"role"`
	jwt.RegisteredClaims
}

// NewJWTManager creates a new JWTManager.
func NewJWTManager(secretKey string, tokenDuration, refreshDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:       secretKey,
		tokenDuration:   tokenDuration,
		refreshDuration: refreshDuration,
	}
}

// GenerateTokens generates both access and refresh JWT tokens for a user.
func (m *JWTManager) GenerateTokens(user *model.User) (accessToken, refreshToken string, err error) {
	// --- デバッグログを追加 ---
	log.Println("--- Generating new JWT token (v2 - with debug logs) ---")
	now := time.Now()
	expTime := now.Add(m.tokenDuration)
	log.Printf("[DEBUG] Current time (iat): %v (%d)\n", now, now.Unix())
	log.Printf("[DEBUG] Expiration time (exp): %v (%d)\n", expTime, expTime.Unix())
	log.Printf("[DEBUG] Token duration from config: %v\n", m.tokenDuration)
	// --- デバッグログここまで ---

	// Generate Access Token
	accessClaims := &UserClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			// ExpiresAt: トークンの有効期限 (現在時刻 + 有効期間)
			ExpiresAt: jwt.NewNumericDate(expTime),
			// IssuedAt: トークンの発行日時 (現在時刻)
			IssuedAt: jwt.NewNumericDate(now),
		},
	}

	// --- デバッグログを追加 ---
	log.Printf("[DEBUG] Claims to be used for access token: %+v\n", accessClaims)
	log.Println("-----------------------------------------------------")
	// --- デバッグログここまで ---

	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(m.secretKey))
	if err != nil {
		return "", "", err
	}

	// Generate Refresh Token
	refreshClaims := &UserClaims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			// Refreshトークンの有効期限
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.refreshDuration)),
			// Refreshトークンの発行日時
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}
	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(m.secretKey))
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// Verify verifies the JWT token and returns user claims if valid.
func (m *JWTManager) Verify(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(m.secretKey), nil
		},
	)

	if err != nil {
		// エラー内容をより詳細にログ出力
		log.Printf("[DEBUG] Token verification failed. Raw Error: %v", err)
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// GetUserIDFromContext retrieves the user ID from the gin context.
func GetUserIDFromContext(c *gin.Context) (uint64, bool) {
	userIDVal, exists := c.Get(UserIDKey)
	if !exists {
		return 0, false
	}

	userID, ok := userIDVal.(uint64)
	return userID, ok
}
