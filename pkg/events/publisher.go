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

// Authentication Events
func (p *Publisher) TrackLogin(ctx context.Context, userID uint64, success bool, metadata map[string]string) error {
	event := models.AnalyticsEvent{
		UserID: userID,
		Event:  models.EventUserLogin,
		Properties: map[string]string{
			"success": strconv.FormatBool(success),
		},
	}
	if metadata != nil {
		for k, v := range metadata {
			event.Properties[k] = v
		}
	}

	// Publish analytics event
	if err := p.PublishAnalytics(ctx, event); err != nil {
		return err
	}

	// Publish audit log
	return p.LogUserAction(ctx, userID, models.ActionLogin, models.ResourceUser, strconv.FormatUint(userID, 10), map[string]interface{}{
		"success":  success,
		"metadata": metadata,
	})
}

func (p *Publisher) TrackLogout(ctx context.Context, userID uint64) error {
	// Publish analytics event
	event := models.AnalyticsEvent{
		UserID: userID,
		Event:  models.EventUserLogout,
	}
	if err := p.PublishAnalytics(ctx, event); err != nil {
		return err
	}

	// Publish audit log
	return p.LogUserAction(ctx, userID, models.ActionLogout, models.ResourceUser, strconv.FormatUint(userID, 10), nil)
}

func (p *Publisher) TrackRegistration(ctx context.Context, userID uint64, metadata map[string]string) error {
	event := models.AnalyticsEvent{
		UserID:     userID,
		Event:      models.EventUserSignup,
		Properties: metadata,
	}
	if err := p.PublishAnalytics(ctx, event); err != nil {
		return err
	}
	// Publish audit log
	return p.LogUserAction(ctx, userID, models.ActionCreate, models.ResourceUser, strconv.FormatUint(userID, 10), map[string]interface{}{
		"metadata": metadata,
	})
}

// Page View Events
func (p *Publisher) TrackPageView(ctx context.Context, userID uint64, page string, metadata map[string]string) error {
	properties := map[string]string{
		"page": page,
	}
	if metadata != nil {
		for k, v := range metadata {
			properties[k] = v
		}
	}

	event := models.AnalyticsEvent{
		UserID:     userID,
		Event:      models.EventPageView,
		Properties: properties,
	}
	return p.PublishAnalytics(ctx, event)
}

// API Request Events
func (p *Publisher) TrackAPIRequest(ctx context.Context, userID uint64, endpoint, method string, statusCode int, metadata map[string]string) error {
	properties := map[string]string{
		"endpoint":    endpoint,
		"method":      method,
		"status_code": strconv.Itoa(statusCode),
	}
	if metadata != nil {
		for k, v := range metadata {
			properties[k] = v
		}
	}

	event := models.AnalyticsEvent{
		UserID:     userID,
		Event:      models.EventAPIRequest,
		Properties: properties,
	}
	return p.PublishAnalytics(ctx, event)
}

// User Action Events
func (p *Publisher) LogUserAction(ctx context.Context, userID uint64, action, resource, resourceID string, details map[string]interface{}) error {
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

// Error Events
func (p *Publisher) TrackError(ctx context.Context, userID uint64, errorType, message string, metadata map[string]string) error {
	properties := map[string]string{
		"error_type": errorType,
		"message":    message,
	}
	if metadata != nil {
		for k, v := range metadata {
			properties[k] = v
		}
	}

	event := models.AnalyticsEvent{
		UserID:     userID,
		Event:      models.EventErrorOccured,
		Properties: properties,
	}
	return p.PublishAnalytics(ctx, event)
}
