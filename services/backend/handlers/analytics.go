package handlers

import (
	"log"
	"net/http"
	"strconv"

	"go-turbo/pkg/database/clickhouse"
	"go-turbo/pkg/models"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	clickhouse *clickhouse.Client
}

func NewAnalyticsHandler(clickhouse *clickhouse.Client) *AnalyticsHandler {
	return &AnalyticsHandler{clickhouse: clickhouse}
}

func (h *AnalyticsHandler) GetEvents(c *gin.Context) {
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

	// Get events from ClickHouse
	events, err := h.clickhouse.GetAnalyticsEvents(c.Request.Context(), uid)
	if err != nil {
		log.Printf("Error fetching analytics events: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch analytics events", "details": err.Error()})
		return
	}

	if events == nil {
		events = []models.AnalyticsEvent{} // Return empty array instead of null
	}

	c.JSON(http.StatusOK, events)
}
