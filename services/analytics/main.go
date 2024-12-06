package main

import (
	"context"
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

	ctx := context.Background()

	// Create analytics table
	if err := clickhouseClient.CreateAnalyticsTable(ctx); err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	// Initialize Gin router
	r := gin.Default()

	// Setup routes
	r.POST("/track", func(c *gin.Context) {
		var event models.AnalyticsEvent
		if err := c.ShouldBindJSON(&event); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if event.Timestamp.IsZero() {
			event.Timestamp = time.Now()
		}

		if err := clickhouseClient.InsertAnalyticsEvent(c.Request.Context(), event); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store event"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// Start server
	port := os.Getenv("ANALYTICS_PORT")
	if port == "" {
		port = "8081"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Analytics service starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down analytics service...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
}
