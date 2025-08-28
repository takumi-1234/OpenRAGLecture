// OpenRAGLecture/internal/domain/model/feedback.go
package model

import "github.com/google/uuid"

// Question represents a user's query log.
type Question struct {
	Base
	QueryID         uuid.UUID `gorm:"type:char(36);not null;index"`
	UserID          uint64
	CourseID        uint64
	SemesterID      uint64
	RawQuery        string `gorm:"type:text;not null"`
	ExpandedQuery   string `gorm:"type:text"`
	AmbiguityScore  float32
	UseLLMExpansion bool  `gorm:"not null;default:false"`
	TracingMeta     JSONB `gorm:"type:json"`

	User     User     `gorm:"foreignKey:UserID"`
	Course   Course   `gorm:"foreignKey:CourseID"`
	Semester Semester `gorm:"foreignKey:SemesterID"`
}

// Answer represents the LLM's response.
type Answer struct {
	Base
	QuestionID        uint64 `gorm:"not null;index"`
	ResponseText      string `gorm:"type:longtext;not null"`
	ResponseModel     string `gorm:"size:128"`
	ResponseParams    JSONB  `gorm:"type:json"`
	ResponseStreamRef string `gorm:"size:255"`

	Question Question `gorm:"foreignKey:QuestionID"`
}

// AnswerSource links an answer to the chunks used to generate it.
type AnswerSource struct {
	Base
	AnswerID         uint64  `gorm:"not null;index"`
	ChunkID          uint64  `gorm:"not null"`
	Score            float32 `gorm:"not null"`
	Rank             int     `gorm:"not null"`
	ExtractedSnippet string  `gorm:"type:text"`

	Answer Answer `gorm:"foreignKey:AnswerID"`
	Chunk  Chunk  `gorm:"foreignKey:ChunkID"`
}

// Feedback represents user feedback on an answer.
type Feedback struct {
	Base
	AnswerID uint64 `gorm:"not null;index"`
	UserID   uint64
	ThumbsUp *bool
	Comment  string `gorm:"type:text"`
	Label    JSONB  `gorm:"type:json"`

	Answer Answer `gorm:"foreignKey:AnswerID"`
	User   User   `gorm:"foreignKey:UserID"`
}
