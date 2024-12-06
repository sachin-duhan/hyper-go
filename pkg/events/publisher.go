package events

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"go-turbo/pkg/models"
	"go-turbo/pkg/queue"
)

type Publisher struct {
	rabbitmq *queue.RabbitMQ
}

func NewPublisher(rabbitmq *queue.RabbitMQ) *Publisher {
	return &Publisher{rabbitmq: rabbitmq}
}

func (p *Publisher) PublishAnalytics(ctx context.Context, event models.AnalyticsEvent) error {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	return p.rabbitmq.Publish("analytics_queue", event)
}

func (p *Publisher) PublishAuditLog(ctx context.Context, log models.AuditLog) error {
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}
	return p.rabbitmq.Publish("audit_logs_queue", log)
}

// Helper functions for common events
func (p *Publisher) TrackLogin(ctx context.Context, userID uint, success bool, metadata map[string]string) error {
	event := models.AnalyticsEvent{
		UserID: userID,
		Event:  models.EventUserLogin,
		Properties: map[string]string{
			"success": strconv.FormatBool(success),
		},
	}
	if metadata != nil {
		event.Properties = metadata
	}
	return p.PublishAnalytics(ctx, event)
}

func (p *Publisher) TrackRegistration(ctx context.Context, userID uint, metadata map[string]string) error {
	event := models.AnalyticsEvent{
		UserID:     userID,
		Event:      models.EventUserSignup,
		Properties: metadata,
	}
	return p.PublishAnalytics(ctx, event)
}

func (p *Publisher) LogUserAction(ctx context.Context, userID uint, action, resource, resourceID string, details map[string]interface{}) error {
	detailsJSON, _ := json.Marshal(details)
	log := models.AuditLog{
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Details:    string(detailsJSON),
	}
	return p.PublishAuditLog(ctx, log)
}
