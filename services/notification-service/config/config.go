package config

import (
	"os"
	"strconv"

	"github.com/rs/zerolog/log"
)

type Config struct {
	App      App
	RabbitMQ RabbitMQ
	SMTP     SMTP
}

type App struct {
	Env string
}

type RabbitMQ struct {
	Host     string
	Port     string
	User     string
	Password string
	VHost    string
}

type SMTP struct {
	Host     string
	Port     int
	User     string
	Password string
}

func LoadConfig() *Config {
	return &Config{
		App: App{
			Env: getEnv("APP_ENV", "development"),
		},
		RabbitMQ: RabbitMQ{
			Host:     getEnv("RABBITMQ_HOST", "localhost"),
			Port:     getEnv("RABBITMQ_PORT", "5672"),
			User:     getEnv("RABBITMQ_USER", "sayur_user"),
			Password: getEnv("RABBITMQ_PASSWORD", "sayur_password"),
			VHost:    getEnv("RABBITMQ_VHOST", "/"),
		},
		SMTP: SMTP{
			Host:     getEnv("SMTP_HOST", "sandbox.smtp.mailtrap.io"),
			Port:     getEnvAsInt("SMTP_PORT", 2525),
			User:     getEnv("SMTP_USER", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	log.Warn().Str("key", key).Str("default", defaultValue).Msg("[Config] Using default value for environment variable")
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		intValue, err := strconv.Atoi(value)
		if err == nil {
			return intValue
		}
		log.Error().Err(err).Str("key", key).Str("value", value).Msg("[Config] Failed to parse environment variable as int")
	}
	log.Warn().Str("key", key).Int("default", defaultValue).Msg("[Config] Using default value for environment variable")
	return defaultValue
}
