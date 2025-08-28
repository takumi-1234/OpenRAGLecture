// OpenRAGLecture/internal/domain/model/user.go
package model

import (
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

// Role defines the user role enumeration.
type Role string

const (
	RoleStudent    Role = "student"
	RoleInstructor Role = "instructor"
	RoleTA         Role = "ta"
	RoleAdmin      Role = "admin"
)

// User represents a user in the system.
type User struct {
	Base
	ExternalID   sql.NullString `gorm:"size:128;index"`
	Email        string         `gorm:"size:255;not null;unique"`
	PasswordHash sql.NullString `gorm:"size:255"`
	DisplayName  sql.NullString `gorm:"size:128"`
	Role         Role           `gorm:"type:enum('student','instructor','ta','admin');not null;default:'student'"`
	IsActive     bool           `gorm:"not null;default:true"`
}

// SetPassword hashes the password and sets it on the user model.
func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = sql.NullString{String: string(hash), Valid: true}
	return nil
}

// CheckPassword compares a plaintext password with the user's hashed password.
func (u *User) CheckPassword(password string) bool {
	if !u.PasswordHash.Valid {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash.String), []byte(password))
	return err == nil
}
