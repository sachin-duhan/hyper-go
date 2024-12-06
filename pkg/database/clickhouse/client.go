package clickhouse

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-turbo/pkg/models"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type Client struct {
	conn driver.Conn
}

func NewClient(host, database, username, password string) (*Client, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{host},
		Auth: clickhouse.Auth{
			Database: database,
			Username: username,
			Password: password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		DialTimeout:     30 * time.Second,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	})
	if err != nil {
		return nil, fmt.Errorf("error connecting to ClickHouse: %w", err)
	}

	// Test connection
	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("error pinging ClickHouse: %w", err)
	}

	return &Client{conn: conn}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) CreateAnalyticsTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS analytics_events (
			id UUID DEFAULT generateUUIDv4(),
			timestamp DateTime64(3),
			user_id UInt64,
			event String,
			metadata String,
			properties Map(String, String)
		)
		ENGINE = MergeTree()
		ORDER BY (timestamp, user_id)
	`
	if err := c.conn.Exec(ctx, query); err != nil {
		return fmt.Errorf("error creating analytics table: %w", err)
	}
	log.Println("Analytics table created/verified successfully")
	return nil
}

func (c *Client) CreateAuditLogsTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS audit_logs (
			id UUID DEFAULT generateUUIDv4(),
			timestamp DateTime64(3),
			user_id UInt64,
			action String,
			resource String,
			resource_id String,
			details String,
			ip_address String,
			user_agent String
		)
		ENGINE = MergeTree()
		ORDER BY (timestamp, user_id)
	`
	if err := c.conn.Exec(ctx, query); err != nil {
		return fmt.Errorf("error creating audit logs table: %w", err)
	}
	log.Println("Audit logs table created/verified successfully")
	return nil
}

func (c *Client) InsertAuditLog(ctx context.Context, log models.AuditLog) error {
	query := `
		INSERT INTO audit_logs (
			timestamp, user_id, action, resource, resource_id,
			details, ip_address, user_agent
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?
		)
	`

	if err := c.conn.Exec(ctx, query,
		log.Timestamp,
		log.UserID,
		log.Action,
		log.Resource,
		log.ResourceID,
		log.Details,
		log.IPAddress,
		log.UserAgent,
	); err != nil {
		return fmt.Errorf("error inserting audit log: %w", err)
	}
	return nil
}

func (c *Client) InsertAnalyticsEvent(ctx context.Context, event models.AnalyticsEvent) error {
	query := `
		INSERT INTO analytics_events (
			timestamp, user_id, event, metadata, properties
		) VALUES (
			?, ?, ?, ?, ?
		)
	`

	if err := c.conn.Exec(ctx, query,
		event.Timestamp,
		event.UserID,
		event.Event,
		event.Metadata,
		event.Properties,
	); err != nil {
		return fmt.Errorf("error inserting analytics event: %w", err)
	}
	return nil
}

func (c *Client) GetAnalyticsEvents(ctx context.Context, userID uint64) ([]models.AnalyticsEvent, error) {
	query := `
		SELECT
			id,
			timestamp,
			user_id,
			event,
			metadata,
			properties
			FROM analytics_events
			WHERE user_id = ?
			ORDER BY timestamp DESC
			LIMIT 1000
	`

	rows, err := c.conn.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying analytics events: %w", err)
	}
	defer rows.Close()

	var events []models.AnalyticsEvent
	for rows.Next() {
		var event models.AnalyticsEvent
		if err := rows.Scan(
			&event.ID,
			&event.Timestamp,
			&event.UserID,
			&event.Event,
			&event.Metadata,
			&event.Properties,
		); err != nil {
			return nil, fmt.Errorf("error scanning analytics event: %w", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating analytics events: %w", err)
	}

	if events == nil {
		events = []models.AnalyticsEvent{} // Return empty array instead of null
	}

	return events, nil
}

func (c *Client) GetAuditLogs(ctx context.Context, userID uint64) ([]models.AuditLog, error) {
	query := `
		SELECT
			id,
			timestamp,
			user_id,
			action,
			resource,
			resource_id,
			details,
			ip_address,
			user_agent
		FROM audit_logs
		WHERE user_id = ?
		ORDER BY timestamp DESC
		LIMIT 1000
	`

	rows, err := c.conn.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying audit logs: %w", err)
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var log models.AuditLog
		if err := rows.Scan(
			&log.ID,
			&log.Timestamp,
			&log.UserID,
			&log.Action,
			&log.Resource,
			&log.ResourceID,
			&log.Details,
			&log.IPAddress,
			&log.UserAgent,
		); err != nil {
			return nil, fmt.Errorf("error scanning audit log: %w", err)
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating audit logs: %w", err)
	}

	if logs == nil {
		logs = []models.AuditLog{} // Return empty array instead of null
	}

	return logs, nil
}
