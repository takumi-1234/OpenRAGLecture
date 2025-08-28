// OpenRAGLecture/internal/domain/model/chunk.go
package model

// Chunk represents a piece of text to be vectorized.
type Chunk struct {
	Base
	PageID                uint64 `gorm:"not null;uniqueIndex:uq_chunk_page_index"`
	DocumentID            uint64 `gorm:"not null"`
	CourseID              uint64 `gorm:"not null;index:idx_chunks_course_semester"`
	SemesterID            uint64 `gorm:"not null;index:idx_chunks_course_semester"`
	ChunkIndex            int    `gorm:"not null;uniqueIndex:uq_chunk_page_index"`
	StartOffset           int
	EndOffset             int
	Text                  string `gorm:"type:longtext;not null"`
	TokenCount            int
	EmbeddingID           string `gorm:"size:128;index"` // Qdrant point ID
	EmbeddingModelVersion string `gorm:"size:64"`
	VectorHash            string `gorm:"size:128"`
	ScoreMeta             JSONB  `gorm:"type:json"`

	Page     Page     `gorm:"foreignKey:PageID"`
	Document Document `gorm:"foreignKey:DocumentID"`
	Course   Course   `gorm:"foreignKey:CourseID"`
	Semester Semester `gorm:"foreignKey:SemesterID"`
}

// RetrievedChunk is a struct holding a chunk and its retrieval score.
type RetrievedChunk struct {
	Chunk Chunk
	Score float32
}
