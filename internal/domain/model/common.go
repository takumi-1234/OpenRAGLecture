// open-rag-lecture/internal/domain/model/common.go

package model

import (
	"database/sql"
	"time"
)

// Base model defines common fields for many tables.
// gorm.Model を参考に、ソフトデリート用の DeletedAt を sql.NullTime にしています。
type Base struct {
	ID uint64 `gorm:"primaryKey;autoIncrement"`
	// ★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★
	// 修正点: `default:current_timestamp` タグを削除。
	// GORMが `CreatedAt` というフィールド名を認識し、レコード作成時に
	// 自動でタイムスタンプを挿入するため、このタグは不要であり、
	// MySQLのバージョンによってはDDLエラーの原因となります。
	// ★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★★
	CreatedAt time.Time    `gorm:"not null"`
	UpdatedAt time.Time    `gorm:"not null"`
	DeletedAt sql.NullTime `gorm:"index"`
}
