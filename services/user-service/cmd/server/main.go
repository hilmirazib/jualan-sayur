package main

import (
	"fmt"
	"log"
	"user-service/config"
	"user-service/internal/adapter/handler"
	"user-service/internal/adapter/middleware"
	"user-service/internal/app"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func main() {
	// Initialize zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Load configuration
	cfg := config.NewConfig()

	// Initialize application
	app, err := app.NewApp(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	// Initialize Echo server
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(middleware.CORSMiddleware())
	e.Use(middleware.LoggerMiddleware())

	// Initialize handlers
	userHandler := handler.NewUserHandler(app.UserService)

	// Public routes (no authentication required)
	public := e.Group("/api/v1")
	public.POST("/auth/signin", userHandler.SignIn)

	// Protected routes (authentication required)
	// Uncomment when you have protected endpoints
	// protected := e.Group("/api/v1", middleware.JWTMiddleware(cfg))
	// protected.GET("/users/profile", userHandler.GetProfile)

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "healthy",
			"service": "user-service",
		})
	})

	// Start server
	serverAddr := fmt.Sprintf(":%s", cfg.App.AppPort)
	log.Printf("🚀 User Service starting on %s", serverAddr)
	log.Printf("📚 API Documentation: http://localhost%s/api/v1", serverAddr)

	if err := e.Start(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
