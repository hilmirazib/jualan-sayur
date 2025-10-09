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

## Testing Configuration and Database

### Test Config and Database Connection
```bash
cd services/user-service
go run test/test_config.go
```

### Test Migration Commands
```bash
cd services/user-service

# Test migrate up (run all migrations)
go run cmd/migrate.go -cmd up

# Test migrate down (rollback last migration)
go run cmd/migrate.go -cmd down

# Test with custom DSN
go run cmd/migrate.go -cmd up -dsn "postgres://user:pass@host:port/db?sslmode=disable"
```

**Expected Results:**
- `migrate up`: No output (success) or error message
- `migrate down`: No output (success) or error message
- Database tables should be created/dropped accordingly

**Expected Output:**
```
=== CONFIG TEST ===
App Port: 8080
App Env: development
JWT Secret: secret
JWT Issuer: secret

=== DATABASE CONFIG ===
Host: localhost
Port: 5432
User: sayur_user
Database: sayur_db
Max Open: 10
Max Idle: 20

=== DATABASE CONNECTION TEST ===
Seeded role: Super Admin
Seeded role: Customer
✅ Database connection successful!
✅ Database ping successful!
✅ All tests passed! Config and database setup is working correctly.
```

### Configuration Files

#### `.env` File
Contains environment variables for the application (copy from `.env.example`):
```env
APP_ENV="development"
APP_PORT="8080"
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=your_db_user
DATABASE_PASSWORD=your_db_password
DATABASE_NAME=your_db_name
DATABASE_MAX_OPEN_CONNECTION=10
DATABASE_MAX_IDLE_CONNECTION=20
JWT_SECRET_KEY="your_jwt_secret"
JWT_ISSUER="your_jwt_issuer"
```

**⚠️ Security Note:** Never commit actual credentials to version control. Use `.env` for local development only.

#### `config/config.go`
- Loads configuration from `.env` file using Viper
- Provides structured config access
- Usage: `cfg := config.NewConfig()`

#### `config/database.go`
- Establishes PostgreSQL connection using GORM
- Automatically runs database seeds
- Usage: `db, err := cfg.ConnectionPostgres()`

## Troubleshooting
- If migration fails, ensure Docker services are running: `docker ps`
- To rebuild services: `docker-compose up --build`
- To stop services: `docker-compose down`