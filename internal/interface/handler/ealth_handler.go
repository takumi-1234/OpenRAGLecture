// internal/interface/handler/health_handler.go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HealthHandler handles health check requests.
type HealthHandler struct {
	db *gorm.DB
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// Check verifies the health of the service, including the database connection.
func (h *HealthHandler) Check(c *gin.Context) {
	// Get the underlying sql.DB object from GORM
	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"details": "failed to get db instance",
		})
		return
	}

	// Ping the database to check for connectivity
	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "error",
			"details": "database connection failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
