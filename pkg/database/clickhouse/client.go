package clickhouse

import (
	"context"
	"fmt"
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
		DialTimeout: 30 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("error connecting to ClickHouse: %w", err)
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
	return c.conn.Exec(ctx, query)
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
	return c.conn.Exec(ctx, query)
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

	return c.conn.Exec(ctx, query,
		log.Timestamp,
		log.UserID,
		log.Action,
		log.Resource,
		log.ResourceID,
		log.Details,
		log.IPAddress,
		log.UserAgent,
	)
}

func (c *Client) InsertAnalyticsEvent(ctx context.Context, event models.AnalyticsEvent) error {
	query := `
		INSERT INTO analytics_events (
			timestamp, user_id, event, metadata, properties
		) VALUES (
			?, ?, ?, ?, ?
		)
	`

	return c.conn.Exec(ctx, query,
		event.Timestamp,
		event.UserID,
		event.Event,
		event.Metadata,
		event.Properties,
	)
}
