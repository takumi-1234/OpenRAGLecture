// OpenRAGLecture/internal/domain/model/document.go
package model

import (
	"database/sql/driver"
	"encoding/json"
)

// DocType defines the document type enumeration.
type DocType string

const (
	DocTypePDF     DocType = "pdf"
	DocTypeSlides  DocType = "slides"
	DocTypeNotes   DocType = "notes"
	DocTypeWebpage DocType = "webpage"
	DocTypeOther   DocType = "other"
)

// JSONB represents a JSON data type for GORM.
type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, &j)
}

// Document represents a lecture material.
type Document struct {
	Base
	CourseID   uint64  `gorm:"not null;index"`
	SemesterID uint64  `gorm:"not null;index"`
	Title      string  `gorm:"size:255;not null"`
	SourceURI  string  `gorm:"size:1024"`
	DocType    DocType `gorm:"type:enum('slides','pdf','notes','webpage','other');default:'pdf'"`
	Version    int     `gorm:"not null;default:1"`
	Checksum   string  `gorm:"size:128"`
	Metadata   JSONB   `gorm:"type:json"`

	Course   Course   `gorm:"foreignKey:CourseID"`
	Semester Semester `gorm:"foreignKey:SemesterID"`
}

// Page represents a single page or section within a document.
type Page struct {
	Base
	DocumentID   uint64 `gorm:"not null;index"`
	PageNumber   int
	SectionTitle string `gorm:"size:255"`
	Language     string `gorm:"size:16"`
	Text         string `gorm:"type:longtext"` // For Full-Text Search
	TokenCount   int

	Document Document `gorm:"foreignKey:DocumentID"`
}
