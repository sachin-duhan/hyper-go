package main

import (
	"log"
	"os"

	"go-turbo/pkg/database"
	"go-turbo/pkg/events"
	"go-turbo/pkg/queue"
	"go-turbo/services/backend/handlers"
	"go-turbo/services/backend/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
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

	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Initialize database
	db, err := database.NewDatabase(database.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
	})
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize RabbitMQ
	rabbitmq, err := queue.NewRabbitMQ(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		logger.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}
	defer rabbitmq.Close()

	// Declare required queues
	if err := rabbitmq.DeclareQueue("analytics_queue"); err != nil {
		logger.Fatal("Failed to declare analytics queue", zap.Error(err))
	}
	if err := rabbitmq.DeclareQueue("audit_logs_queue"); err != nil {
		logger.Fatal("Failed to declare audit logs queue", zap.Error(err))
	}

	// Initialize event publisher
	publisher := events.NewPublisher(rabbitmq)

	// Initialize router
	r := gin.Default()

	// Add CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60, // 12 hours
	}))

	// Initialize handlers and middleware
	authHandler := handlers.NewAuthHandler(db, publisher)
	authMiddleware := middleware.NewAuthMiddleware(db)

	// Public routes
	r.POST("/api/auth/login", authHandler.Login)
	r.POST("/api/auth/register", authHandler.Register)

	// Protected routes
	authorized := r.Group("/api")
	authorized.Use(authMiddleware.RequireAuth())
	{
		// Admin routes
		admin := authorized.Group("/admin")
		admin.Use(authMiddleware.RequireRole([]string{"admin"}))
		{
			admin.GET("/users", authHandler.GetUsers)
		}

		// User routes
		user := authorized.Group("/user")
		{
			user.GET("/profile", authHandler.GetProfile)
		}
	}

	// Start server
	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("Server starting", zap.String("port", port))
	if err := r.Run(":" + port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
