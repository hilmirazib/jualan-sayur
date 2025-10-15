package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
	"user-service/config"
	"user-service/internal/adapter/handler"
	"user-service/internal/adapter/message"
	"user-service/internal/adapter/middleware"
	"user-service/internal/adapter/repository"
	"user-service/internal/core/port"
	"user-service/internal/core/service"
	"user-service/utils"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
	"gorm.io/gorm"
)

// App holds all dependencies
type App struct {
	UserService      port.UserServiceInterface
	UserRepo         port.UserRepositoryInterface
	JWTUtil          port.JWTInterface
	DB               *gorm.DB
	RabbitMQChannel  *amqp.Channel
	// Add other services here as they are created
}

// checkPortAvailability checks if the given port is available
func checkPortAvailability(port string) error {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	ln.Close()
	return nil
}

// RunServer starts the HTTP server with graceful shutdown
func RunServer() {
	// Initialize zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Load configuration
	cfg := config.NewConfig()

	// Check if APP_PORT is available
	if err := checkPortAvailability(cfg.App.AppPort); err != nil {
		log.Fatalf("Port %s is not available: %v", cfg.App.AppPort, err)
	}

	// Initialize application
	app, err := NewApp(cfg)
	if err != nil {
		log.Printf("⚠️  Database not available, starting server in offline mode: %v", err)
		log.Printf("🚀 Server will start but API endpoints will return database errors")
		log.Printf("💡 To fix: Start PostgreSQL and run migrations")
		// Continue with nil app - handlers will handle nil gracefully
	}

	// Initialize Echo server
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(middleware.CORSMiddleware())
	e.Use(middleware.LoggerMiddleware())

	// Initialize repositories
	redisClient := cfg.RedisClient()
	sessionRepo := repository.NewSessionRepository(redisClient, cfg)
	verificationTokenRepo := repository.NewVerificationTokenRepository(app.DB)

	// Initialize message publishers
	emailPublisher := message.NewEmailPublisher(app.RabbitMQChannel)

	// Initialize services with new dependencies
	app.UserService = service.NewUserService(app.UserRepo, sessionRepo, app.JWTUtil, verificationTokenRepo, emailPublisher, cfg)

	// Initialize handlers
	userHandler := handler.NewUserHandler(app.UserService)

	// Public routes (no authentication required)
	public := e.Group("/api/v1")
	public.POST("/auth/signin", userHandler.SignIn)
	public.POST("/auth/signup", userHandler.CreateUserAccount)
	public.GET("/auth/verify", userHandler.VerifyUserAccount)

	// Protected routes (authentication required)
	admin := e.Group("/api/v1/admin", middleware.JWTMiddleware(cfg, sessionRepo))
	admin.GET("/check", userHandler.AdminCheck)

	// Root endpoint - redirect to health
	e.GET("/", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"message": "User Service API",
			"version": "1.0.0",
			"health": "/health",
			"docs": "/api/v1",
		})
	})

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "healthy",
			"service": "user-service",
		})
	})

	// Start server in a goroutine
	go func() {
		serverAddr := fmt.Sprintf(":%s", cfg.App.AppPort)
		log.Printf("🚀 User Service starting on %s", serverAddr)
		log.Printf("📚 API Documentation: http://localhost%s/api/v1", serverAddr)

		if err := e.Start(serverAddr); err != nil {
			log.Printf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests 5 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

// NewApp initializes the application with all dependencies
func NewApp(cfg *config.Config) (*App, error) {
	// Initialize database connection
	db, err := cfg.ConnectionPostgres()
	if err != nil {
		log.Fatalf("[RunServer-1] %v", err)
		return nil, err
	}

	// Initialize Redis client
	redisClient := cfg.RedisClient()

	// Initialize RabbitMQ connection
	rabbitMQChannel, err := cfg.ConnectionRabbitMQ()
	if err != nil {
		log.Printf("⚠️  RabbitMQ not available: %v", err)
		log.Printf("💡 Email verification will not work until RabbitMQ is started")
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.DB)
	sessionRepo := repository.NewSessionRepository(redisClient, cfg)

	// Initialize utilities
	jwtUtil := utils.NewJWTUtil(cfg)

	// Initialize message publishers
	var emailPublisher port.EmailInterface
	if rabbitMQChannel != nil {
		emailPublisher = message.NewEmailPublisher(rabbitMQChannel)
	}

	// Initialize services
	userService := service.NewUserService(userRepo, sessionRepo, jwtUtil, nil, emailPublisher, cfg)

	return &App{
		UserService:     userService,
		UserRepo:        userRepo,
		JWTUtil:         jwtUtil,
		DB:              db.DB,
		RabbitMQChannel: rabbitMQChannel,
	}, nil
}

// Example usage function
func (a *App) ExampleUsage() {
	// This is just an example of how to use the service
	// In a real application, this would be called from handlers

	// ctx := context.Background()
	// user, err := a.UserService.GetUserByEmail(ctx, "user@example.com")
	// if err != nil {
	//     log.Error().Err(err).Msg("Failed to get user")
	//     return
	// }
	// fmt.Printf("User: %+v\n", user)
}
