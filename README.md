# MICRO-SAYUR Project Setup Guide

## Prerequisites
- Docker & Docker Compose installed
- Go 1.21+ (for local development)

## Quick Setup

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd MICRO-SAYUR
   ```

2. **Start all services with Docker**
   ```bash
   docker-compose up -d
   ```
   This will start:
   - PostgreSQL (port 5432)
   - Redis (port 6379)
   - RabbitMQ (port 5672, management on 15672)
   - Adminer (port 8080) - Web-based database management

3. **Run database migrations**
   ```bash
   cd services/user-service
   go mod tidy
   go run cmd/migrate.go -cmd up
   ```

4. **Access the database visually**
   - Open browser: http://localhost:8080
   - System: PostgreSQL
   - Server: sayur-postgres
   - Username: sayur_user
   - Password: sayur_password
   - Database: sayur_db

## Database Tables Created
After migration, the following tables will be available:
- `users` - User accounts
- `roles` - User roles
- `user_role` - User-role relationships
- `schema_migrations` - Migration tracking

## Services Access
- **Adminer (Database UI)**: http://localhost:8080
- **RabbitMQ Management**: http://localhost:15672 (user: sayur_user, pass: sayur_password)

## Troubleshooting
- If migration fails, ensure Docker services are running: `docker ps`
- To rebuild services: `docker-compose up --build`
- To stop services: `docker-compose down`