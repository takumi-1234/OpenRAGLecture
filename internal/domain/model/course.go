// OpenRAGLecture/internal/domain/model/course.go
package model

import "time"

// Semester represents an academic semester.
type Semester struct {
	Base
	Name      string    `gorm:"size:64;not null"`
	StartDate time.Time `gorm:"not null"`
	EndDate   time.Time `gorm:"not null"`
	Note      string    `gorm:"size:255"`
}

// Course represents a lecture course.
type Course struct {
	Base
	Code         string `gorm:"size:50;not null;index"`
	Title        string `gorm:"size:255;not null"`
	SemesterID   uint64 `gorm:"not null;index"`
	InstructorID uint64
	Description  string `gorm:"type:text"`
	IsActive     bool   `gorm:"not null;default:true"`

	Semester   Semester `gorm:"foreignKey:SemesterID"`
	Instructor User     `gorm:"foreignKey:InstructorID"`
}

// Enrollment represents a user's enrollment in a course.
type Enrollment struct {
	Base
	UserID   uint64 `gorm:"not null;uniqueIndex:uq_enrollment_user_course"`
	CourseID uint64 `gorm:"not null;uniqueIndex:uq_enrollment_user_course"`
	Role     string `gorm:"type:enum('student','auditor','ta');not null;default:'student'"`

	User   User
	Course Course
}
