package app

import (
	"user-service/config"
	"user-service/internal/adapter/repository"
	"user-service/internal/core/port"
	"user-service/internal/core/service"
)

// App holds all dependencies
type App struct {
	UserService port.UserServiceInterface
	// Add other services here as they are created
}

func RunServer() {
	
}

// NewApp initializes the application with all dependencies
func NewApp(cfg *config.Config) (*App, error) {
	// Initialize database connection
	db, err := cfg.ConnectionPostgres()
	if err != nil {
		return nil, err
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.DB)

	// Initialize services
	userService := service.NewUserService(userRepo, cfg)

	return &App{
		UserService: userService,
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
