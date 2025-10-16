package main

import (
	"context"
	"notification-service/config"
	"notification-service/internal/adapter/consumer"
	"notification-service/internal/core/service"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
)

func main() {
	// Initialize logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Load configuration
	cfg := config.LoadConfig()
	logger.Info().Str("env", cfg.App.Env).Msg("Starting notification service")

	// Connect to RabbitMQ
	connString := "amqp://" + cfg.RabbitMQ.User + ":" + cfg.RabbitMQ.Password + "@" + cfg.RabbitMQ.Host + ":" + cfg.RabbitMQ.Port + cfg.RabbitMQ.VHost
	conn, err := amqp.Dial(connString)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to RabbitMQ")
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to open channel")
	}
	defer channel.Close()

	logger.Info().Msg("Connected to RabbitMQ")

	// Initialize services
	emailService := service.NewEmailService(cfg)

	// Initialize consumer
	emailConsumer := consumer.NewEmailConsumer(cfg, emailService, channel)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start consuming messages
	if err := emailConsumer.StartConsuming(ctx); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start consuming")
	}

	logger.Info().Msg("Notification service started successfully")

	// Wait for shutdown signal
	<-sigChan
	logger.Info().Msg("Shutting down notification service...")

	// Cancel context to stop consumer
	cancel()

	logger.Info().Msg("Notification service stopped")
}
