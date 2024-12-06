package middleware

import (
	"go-turbo/pkg/events"

	"github.com/gin-gonic/gin"
)

type AnalyticsMiddleware struct {
	publisher *events.Publisher
}

func NewAnalyticsMiddleware(publisher *events.Publisher) *AnalyticsMiddleware {
	return &AnalyticsMiddleware{publisher: publisher}
}

func (m *AnalyticsMiddleware) TrackRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request
		c.Next()

		// Get user ID if authenticated
		if id, exists := c.Get("userID"); exists {
			userID := id.(uint64)
			// Track API request
			m.publisher.TrackAPIRequest(
				c.Request.Context(),
				userID,
				c.Request.URL.Path,
				c.Request.Method,
				c.Writer.Status(),
				map[string]string{
					"ip_address": c.ClientIP(),
					"user_agent": c.Request.UserAgent(),
				},
			)
		}
	}
}

func (m *AnalyticsMiddleware) TrackPageView() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID if authenticated
		if id, exists := c.Get("userID"); exists {
			userID := id.(uint64)
			// Track page view
			m.publisher.TrackPageView(
				c.Request.Context(),
				userID,
				c.Request.URL.Path,
				map[string]string{
					"referrer":   c.Request.Referer(),
					"ip_address": c.ClientIP(),
					"user_agent": c.Request.UserAgent(),
				},
			)
		}

		c.Next()
	}
}
