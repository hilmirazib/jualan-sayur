# User Service

User Service adalah microservice untuk manajemen autentikasi dan user menggunakan Clean Architecture pattern dengan Go.

## 🚀 Quick Start

### 1. Setup Environment

```bash
# Copy environment file
cp .env.example .env

# Edit .env file dengan konfigurasi database Anda
```

### 2. Setup Database

```bash
# Jalankan migrations
go run cmd/migrate.go up

# Optional: Jalankan seeds untuk data dummy
go run cmd/migrate.go seed
```

### 3. Start Server

```bash
# Jalankan server
go run ./cmd/server

# Server akan start di port 8080 (default)
```

### 4. Health Check

```bash
curl http://localhost:8080/health
```

## 📚 API Documentation

### Sign In

**Endpoint:** `POST /api/v1/auth/signin`

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Success Response (200):**
```json
{
  "message": "Sign in successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "role": "user",
    "id": 1,
    "name": "John Doe",
    "email": "user@example.com",
    "phone": "+628123456789",
    "lat": "-6.2088",
    "lng": "106.8456"
  }
}
```

**Error Responses:**

**400 Bad Request - Invalid Input:**
```json
{
  "message": "Email and password are required"
}
```

**404 Not Found - User Not Found:**
```json
{
  "message": "User not found"
}
```

**401 Unauthorized - Wrong Password:**
```json
{
  "message": "Incorrect password"
}
```

## 🧪 Testing

### Unit Tests

```bash
# Run all tests
go test ./...

# Run service tests only
go test ./internal/core/service/

# Run with verbose output
go test -v ./internal/core/service/
```

### API Testing dengan cURL

#### 1. Sign In - Success
```bash
curl -X POST http://localhost:8080/api/v1/auth/signin \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

#### 2. Sign In - Invalid Email
```bash
curl -X POST http://localhost:8080/api/v1/auth/signin \
  -H "Content-Type: application/json" \
  -d '{
    "email": "",
    "password": "password123"
  }'
```

#### 3. Sign In - User Not Found
```bash
curl -X POST http://localhost:8080/api/v1/auth/signin \
  -H "Content-Type: application/json" \
  -d '{
    "email": "notfound@example.com",
    "password": "password123"
  }'
```

#### 4. Sign In - Wrong Password
```bash
curl -X POST http://localhost:8080/api/v1/auth/signin \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "wrongpassword"
  }'
```

### API Testing dengan Postman

1. **Import Collection:**
   - Buat collection baru di Postman
   - Tambahkan request baru dengan method POST
   - URL: `http://localhost:8080/api/v1/auth/signin`
   - Headers: `Content-Type: application/json`
   - Body: raw JSON seperti contoh di atas

2. **Test Scenarios:**
   - ✅ Valid credentials → Should return 200 with JWT token
   - ❌ Empty email → Should return 400
   - ❌ User not found → Should return 404
   - ❌ Wrong password → Should return 401

### Using JWT Token

Setelah mendapat token dari Sign In, gunakan untuk API yang protected:

```bash
curl -X GET http://localhost:8080/api/v1/protected-endpoint \
  -H "Authorization: Bearer YOUR_JWT_TOKEN_HERE"
```

## 🏗️ Architecture

```
HTTP Request
    ↓
Handler Layer (Echo)
    ↓
Service Layer (Business Logic)
    ↓
Repository Layer (Data Access)
    ↓
Database (PostgreSQL)
```

### Layers:

- **Handler**: HTTP request/response handling
- **Service**: Business logic, validation, JWT generation
- **Repository**: Database operations
- **Entity**: Domain models
- **Port**: Interface definitions

## 🔧 Configuration

### Environment Variables (.env)

```env
# App Configuration
APP_NAME=user-service
APP_ENV=development
APP_PORT=8080
JWT_SECRET_KEY=your-super-secret-jwt-key-here
JWT_ISSUER=user-service

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=user_service
DB_MAX_OPEN_CONNS=10
DB_MAX_IDLE_CONNS=5
```

## 📦 Dependencies

- **Echo**: Web framework
- **GORM**: ORM untuk database
- **PostgreSQL**: Database
- **JWT**: Token authentication
- **Zerolog**: Structured logging
- **Testify**: Testing framework

## 🐳 Docker

```bash
# Build image
docker build -t user-service .

# Run container
docker run -p 8080:8080 --env-file .env user-service
```

## 📝 Development

### Adding New Features

1. **Repository**: Tambahkan method di interface dan implementasi
2. **Service**: Tambahkan business logic
3. **Handler**: Tambahkan HTTP endpoint
4. **Tests**: Tambahkan unit tests

### Code Structure

```
services/user-service/
├── cmd/                    # Application entrypoints
├── config/                 # Configuration
├── database/               # Migrations & seeds
├── internal/
│   ├── adapter/           # External interfaces
│   │   ├── handler/       # HTTP handlers
│   │   ├── middleware/    # HTTP middleware
│   │   └── repository/    # Data repositories
│   ├── app/               # Application setup
│   └── core/              # Business logic
│       ├── domain/        # Domain models
│       ├── port/          # Interfaces
│       └── service/       # Business services
├── utils/                  # Utilities
└── mocks/                  # Test mocks
```

## 🤝 Contributing

1. Fork repository
2. Create feature branch
3. Add tests untuk perubahan
4. Pastikan semua tests pass
5. Submit pull request

## 📄 License

This project is licensed under the MIT License.
