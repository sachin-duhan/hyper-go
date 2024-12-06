package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-turbo/pkg/database/clickhouse"
	"go-turbo/pkg/models"
	"go-turbo/pkg/queue"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	// Set Gin mode based on environment
	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize RabbitMQ
	rabbitmq, err := queue.NewRabbitMQ(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitmq.Close()

	// Initialize ClickHouse
	clickhouseClient, err := clickhouse.NewClient(
		os.Getenv("CLICKHOUSE_HOST"),
		os.Getenv("CLICKHOUSE_DATABASE"),
		os.Getenv("CLICKHOUSE_USER"),
		os.Getenv("CLICKHOUSE_PASSWORD"),
	)
	if err != nil {
		log.Fatalf("Failed to connect to ClickHouse: %v", err)
	}
	defer clickhouseClient.Close()

	// Create context with timeout for initialization
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create audit logs table
	if err := clickhouseClient.CreateAuditLogsTable(ctx); err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	// Declare queue
	if err := rabbitmq.DeclareQueue("audit_logs_queue"); err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	// Start consuming messages
	msgs, err := rabbitmq.Consume("audit_logs_queue")
	if err != nil {
		log.Fatalf("Failed to consume queue: %v", err)
	}

	// Initialize Gin router
	r := gin.Default()

	// Setup routes
	r.POST("/audit", func(c *gin.Context) {
		var auditLog models.AuditLog
		if err := c.ShouldBindJSON(&auditLog); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if auditLog.Timestamp.IsZero() {
			auditLog.Timestamp = time.Now()
		}

		if err := clickhouseClient.InsertAuditLog(c.Request.Context(), auditLog); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store audit log"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// Start server
	port := os.Getenv("AUDIT_LOGS_PORT")
	if port == "" {
		port = "8082"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start message processor
	go processMessages(msgs, clickhouseClient)

	// Start server
	go func() {
		log.Printf("Audit logs service starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down audit logs service...")

	// Shutdown HTTP server
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
}

func processMessages(msgs <-chan queue.Message, client *clickhouse.Client) {
	for msg := range msgs {
		var auditLog models.AuditLog
		if err := json.Unmarshal(msg.Body, &auditLog); err != nil {
			log.Printf("Error parsing message: %v", err)
			msg.Nack(false) // Negative acknowledgment, don't requeue
			continue
		}

		// Set timestamp if not provided
		if auditLog.Timestamp.IsZero() {
			auditLog.Timestamp = time.Now()
		}

		// Insert audit log into ClickHouse
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := client.InsertAuditLog(ctx, auditLog); err != nil {
			log.Printf("Error inserting audit log: %v", err)
			msg.Nack(true) // Negative acknowledgment, requeue
			cancel()
			continue
		}
		cancel()

		// Acknowledge message
		msg.Ack()
		log.Printf("Processed audit log: action=%s, userID=%d, resource=%s",
			auditLog.Action, auditLog.UserID, auditLog.Resource)
	}
}
