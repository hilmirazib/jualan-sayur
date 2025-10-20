# MICRO-SAYUR Makefile
# Commands for development, testing, and deployment

.PHONY: help setup test clean docker-up docker-down setup-test-data

# Default target
help:
	@echo "Available commands:"
	@echo "  setup          - Setup development environment"
	@echo "  test           - Run all tests"
	@echo "  clean          - Clean build artifacts"
	@echo "  docker-up      - Start all Docker services"
	@echo "  docker-down    - Stop all Docker services"
	@echo "  setup-test-data- Setup test data and Hoppscotch collections"
	@echo "  migrate-up     - Run database migrations"
	@echo "  migrate-down   - Rollback database migrations"

# Setup development environment
setup:
	@echo "Setting up development environment..."
	docker-compose up -d postgres redis rabbitmq
	@echo "Waiting for services to be ready..."
	sleep 10
	cd services/user-service && go mod tidy
	@echo "Development environment ready!"

# Start all Docker services
docker-up:
	docker-compose up -d

# Stop all Docker services
docker-down:
	docker-compose down

# Setup test data for API testing
setup-test-data:
	@echo "📋 API Testing Setup Instructions:"
	@echo ""
	@echo "✅ Collections ready at: scripts/hoppscotch-web-collections.json"
	@echo ""
	@echo "🎯 Next steps:"
	@echo "1. Open Hoppscotch Web: https://hoppscotch.io"
	@echo "2. Import collection: scripts/hoppscotch-web-collections.json"
	@echo "3. Register/Login to get JWT token"
	@echo "4. Update JWT_TOKEN environment variable"
	@echo "5. Start testing APIs!"

# Run database migrations
migrate-up:
	cd services/user-service && go run cmd/migrate.go -cmd up

# Rollback database migrations
migrate-down:
	cd services/user-service && go run cmd/migrate.go -cmd down

# Run tests
test:
	cd services/user-service && go test ./...

# Clean build artifacts
clean:
	docker-compose down -v
	find . -name "*.log" -delete
	find . -name "*.tmp" -delete

# Quick development setup
dev-setup: docker-up migrate-up setup-test-data
	@echo ""
	@echo "🎉 Development environment ready!"
	@echo "🌐 Hoppscotch Web: https://hoppscotch.io (import scripts/hoppscotch-web-collections.json)"
	@echo "📊 Adminer: http://localhost:8081"
	@echo "🐰 RabbitMQ: http://localhost:15672"
