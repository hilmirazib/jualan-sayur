# User Service

## Overview
User Service adalah microservice yang menangani semua operasi terkait manajemen pengguna dalam platform MICRO-SAYUR. Service ini bertanggung jawab atas autentikasi, autorisasi, dan manajemen profil pengguna.

## Features

### ✅ Implemented Features
- **User Registration**: Pendaftaran user baru dengan email verification
- **User Authentication**: Login dengan JWT token
- **User Authorization**: Role-based access control (RBAC)
- **Profile Management**:
  - Get user profile data
  - Update user profile data (name, email, phone, address, location, photo)
  - Upload profile image
- **Role Management**:
  - Get all roles with optional search (Super Admin only)
  - Get role by ID with associated users (Super Admin only)
  - Create new roles with validation (Super Admin only)
  - Role-based permissions with Super Admin access control
- **Email Verification**: Verifikasi email untuk aktivasi akun
- **Password Reset**: Forgot password dengan email reset link
- **Session Management**: Manajemen session dengan Redis

### 🚧 Planned Features
- Social login (Google, Facebook)
- Two-factor authentication (2FA)
- User preferences
- Account deletion
- Audit logging

## Architecture

### Clean Architecture (Hexagonal)
```
internal/
├── core/
│   ├── domain/          # Business entities & rules
│   ├── service/         # Application use cases
│   └── port/            # Interfaces/contracts
└── adapter/
    ├── handler/         # HTTP handlers
    ├── repository/      # Data access
    ├── middleware/      # Cross-cutting concerns
    └── message/         # Message publishers
```

## API Endpoints

### Authentication
```
POST   /api/v1/auth/register          # Register new user
POST   /api/v1/auth/login             # User login
POST   /api/v1/auth/logout            # User logout
POST   /api/v1/auth/refresh           # Refresh JWT token
POST   /api/v1/auth/forgot-password   # Request password reset
POST   /api/v1/auth/reset-password    # Reset password with token
```

### User Management
```
GET    /api/v1/auth/profile           # Get user profile
PUT    /api/v1/auth/profile           # Update user profile
POST   /api/v1/auth/profile/image-upload  # Upload profile image
GET    /api/v1/users/:id              # Get user by ID (admin)
PUT    /api/v1/users/:id              # Update user (admin)
DELETE /api/v1/users/:id              # Delete user (admin)
```

### Email Verification
```
POST   /api/v1/verification/send       # Send verification email
GET    /api/v1/verification/verify    # Verify email token
```

### Admin (Role-based)
```
GET    /api/v1/admin/users            # List all users
GET    /api/v1/admin/users/:id        # Get user details
PUT    /api/v1/admin/users/:id/role   # Assign user role
GET    /api/v1/admin/roles            # Get all roles with search (Super Admin only)
POST   /api/v1/admin/roles            # Create new role (Super Admin only)
GET    /api/v1/admin/roles/:id        # Get role by ID with users (Super Admin only)
```

## Database Schema

### Tables
- `users` - User account information
- `roles` - User roles (admin, user, etc.)
- `user_roles` - Many-to-many user-role relationships
- `verification_users` - Email verification tokens
- `sessions` - User sessions

### Key Relationships
```
users (1) ──── (N) user_roles (N) ──── (1) roles
   │                                        │
   │                                        │
   ↓                                        │
   └── (1) verification_users (N)           │
   │                                        │
   └── (1) sessions (N)                     │
```

## Dependencies

### External Services
- **PostgreSQL**: Primary database
- **Redis**: Session store & caching
- **RabbitMQ**: Message queue for email notifications

### Internal Dependencies
- **Notification Service**: Email sending via message queue

## Configuration

### Environment Variables
```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=micro_sayur
DB_USER=micro_sayur
DB_PASSWORD=password

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# RabbitMQ
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRATION=24h

# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
```

## Development

### Prerequisites
- Go 1.21+
- PostgreSQL 15+
- Redis 7+
- RabbitMQ 3.12+

### Setup
```bash
# Clone repository
git clone https://github.com/hilmirazib/jualan-sayur.git
cd services/user-service

# Install dependencies
go mod download

# Copy environment file
cp .env.example .env

# Run database migrations
make migrate-up

# Run seeders
make seed

# Run service
make run
```

### Testing
```bash
# Run unit tests
make test

# Run integration tests
make test-integration

# Run with coverage
make test-coverage
```

### Available Commands
```bash
make run              # Start the service
make build            # Build binary
make test             # Run tests
make migrate-up       # Run migrations
make migrate-down     # Rollback migrations
make seed             # Run seeders
make docker-build     # Build Docker image
make docker-run       # Run with Docker
```

## Deployment

### Docker
```bash
# Build image
docker build -t micro-sayur/user-service .

# Run container
docker run -p 8080:8080 micro-sayur/user-service
```

### Docker Compose (Development)
```yaml
version: '3.8'
services:
  user-service:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
      - RABBITMQ_HOST=rabbitmq
    depends_on:
      - postgres
      - redis
      - rabbitmq
```

## Monitoring & Health Checks

### Health Endpoints
```
GET  /health     # Overall health status
GET  /ready      # Readiness probe
GET  /metrics    # Prometheus metrics
```

### Key Metrics
- Request latency
- Error rates
- Database connection pool
- Redis hit/miss ratio
- Message queue throughput

## Security

### Authentication
- JWT tokens with expiration
- Refresh token rotation
- Secure password hashing (bcrypt)

### Authorization
- Role-based access control
- Route-level permissions
- Middleware validation

### Data Protection
- Input validation
- SQL injection prevention
- XSS protection
- Rate limiting

## Error Handling

### Error Codes
- `400` - Bad Request (validation errors)
- `401` - Unauthorized (invalid credentials)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found
- `409` - Conflict (duplicate data)
- `422` - Unprocessable Entity
- `500` - Internal Server Error

### Error Response Format
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": [
      {
        "field": "email",
        "message": "Email format is invalid"
      }
    ]
  }
}
```

## Logging

### Log Levels
- `DEBUG` - Detailed debug information
- `INFO` - General information
- `WARN` - Warning messages
- `ERROR` - Error conditions

### Structured Logging
```json
{
  "level": "INFO",
  "timestamp": "2025-10-16T13:55:21Z",
  "service": "user-service",
  "method": "POST /api/v1/auth/login",
  "user_id": "uuid",
  "duration_ms": 150,
  "status_code": 200
}
```

## Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/new-feature`)
3. Commit changes (`git commit -am 'Add new feature'`)
4. Push to branch (`git push origin feature/new-feature`)
5. Create Pull Request

## API Documentation

### OpenAPI Specification
API documentation tersedia di `/docs` endpoint ketika service running dalam mode development.

### Postman Collection
Collection Postman tersedia di `docs/postman/` directory.
