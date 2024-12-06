package handlers

import (
	"log"
	"net/http"
	"strconv"

	"go-turbo/pkg/database/clickhouse"
	"go-turbo/pkg/models"

	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	clickhouse *clickhouse.Client
}

func NewAuditHandler(clickhouse *clickhouse.Client) *AuditHandler {
	return &AuditHandler{clickhouse: clickhouse}
}

func (h *AuditHandler) GetLogs(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		// If no user_id provided, get current user's ID from context
		if id, exists := c.Get("userID"); exists {
			userID = strconv.FormatUint(id.(uint64), 10)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found"})
			return
		}
	}

	// Convert userID to uint64
	uid, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		log.Printf("Error parsing user ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get logs from ClickHouse
	logs, err := h.clickhouse.GetAuditLogs(c.Request.Context(), uid)
	if err != nil {
		log.Printf("Error fetching audit logs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch audit logs", "details": err.Error()})
		return
	}

	if logs == nil {
		logs = []models.AuditLog{} // Return empty array instead of null
	}

	c.JSON(http.StatusOK, logs)
}
