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
# Build aplikasi CLI
go build -o sayur-api cmd/server/main.go

# Jalankan migrations
./sayur-api migrate up

# Optional: Jalankan seeds untuk data dummy
./sayur-api migrate seed
```

### 3. Start Server

```bash
# Jalankan server dengan CLI
./sayur-api start

# Atau dengan custom port
./sayur-api start --port 3000

# Server akan start di port 8080 (default)
```

### 4. Health Check

```bash
curl http://localhost:8080/health
```

## 🛠️ CLI Commands

Aplikasi ini menggunakan CLI berbasis Cobra untuk kemudahan penggunaan. Berikut adalah command yang tersedia:

### Build Aplikasi

```bash
# Build executable
go build -o sayur-api cmd/server/main.go

# Untuk Windows: sayur-api.exe
# Untuk Linux/Mac: sayur-api (tanpa ekstensi)
```

### Command Utama

#### 1. Start Server
```bash
# Start server dengan konfigurasi default
./sayur-api start

# Start dengan port custom
./sayur-api start --port 3000

# Start dengan environment custom
./sayur-api start --env production

# Kombinasi flags
./sayur-api start --port 3000 --env development --verbose
```

#### 2. Database Migration
```bash
# Jalankan semua migration
./sayur-api migrate up

# Rollback migration terakhir
./sayur-api migrate down

# Jalankan seeding saja
./sayur-api migrate seed

# Custom migration directory
./sayur-api migrate --dir ./custom/migrations up
```

#### 3. Konfigurasi
```bash
# Lihat konfigurasi aktif
./sayur-api config

# Validasi koneksi database
./sayur-api config --validate-db
```

#### 4. Help & Version
```bash
# Lihat semua command
./sayur-api --help

# Lihat help spesifik command
./sayur-api start --help
./sayur-api migrate --help
./sayur-api config --help

# Lihat versi aplikasi
./sayur-api --version
```

### Global Flags

```bash
# Gunakan config file custom
./sayur-api --config ./custom.env start

# Enable verbose output
./sayur-api --verbose start
```

## 📋 Prerequisites

### System Requirements

- Go 1.19+ (untuk development)
- PostgreSQL 12+
- Environment file (`.env`) dengan konfigurasi database

### Untuk Menjalankan Executable

#### **Windows:**
1. **Buka Command Prompt atau PowerShell:**
   - Tekan `Win + R`, ketik `cmd`, tekan Enter
   - Atau cari "Command Prompt" di Start Menu

2. **Navigate ke folder project:**
   ```cmd
   cd C:\path\to\your\project\services\user-service
   ```

3. **Jalankan executable:**
   ```cmd
   sayur-api.exe --version
   sayur-api.exe --help
   sayur-api.exe start
   ```

#### **Linux/Mac:**
1. **Buka Terminal:**
   - Linux: Cari "Terminal" di aplikasi
   - Mac: Tekan `Cmd + Space`, cari "Terminal"

2. **Navigate ke folder project:**
   ```bash
   cd /path/to/your/project/services/user-service
   ```

3. **Berikan permission dan jalankan:**
   ```bash
   chmod +x sayur-api
   ./sayur-api --version
   ./sayur-api --help
   ./sayur-api start
   ```

### ⚠️ **PENTING: Jangan Jalankan di Browser!**

Ketika Anda klik file `sayur-api.exe` langsung dari File Explorer:
- Windows akan bertanya "Buka dengan aplikasi apa?"
- Pilih "Command Prompt" atau "Windows PowerShell"
- Atau ikuti langkah manual di atas

**File executable (.exe) bukan file web yang bisa dibuka di browser!**

## 📚 API Documentation

### Sign Up (Create User Account)

**Endpoint:** `POST /api/v1/auth/signup`

**Request Body:**
```json
{
  "email": "user@example.com",
  "name": "John Doe",
  "password": "password123",
  "password_confirmation": "password123"
}
```

**Success Response (201):**
```json
{
  "message": "Account created successfully. Please check your email for verification.",
  "data": {
    "id": 1,
    "name": "John Doe",
    "email": "user@example.com"
  }
}
```

**Error Responses:**

**422 Unprocessable Entity - Validation Failed:**
```json
{
  "message": "Validation failed",
  "data": null
}
```

**409 Conflict - Email Already Exists:**
```json
{
  "message": "Email already exists",
  "data": null
}
```

**500 Internal Server Error:**
```json
{
  "message": "Internal server error",
  "data": null
}
```

### Verify User Account

**Endpoint:** `GET /api/v1/auth/verify?token=:token`

**Query Parameters:**
- `token`: Verification token received via email

**Success Response (200):**
```json
{
  "message": "Account verified successfully. You can now sign in.",
  "data": null
}
```

**Error Responses:**

**400 Bad Request - Invalid Token:**
```json
{
  "message": "Invalid or expired verification token",
  "data": null
}
```

**500 Internal Server Error:**
```json
{
  "message": "Internal server error",
  "data": null
}
```

### Forgot Password

**Endpoint:** `POST /api/v1/auth/forgot-password`

**Request Body:**
```json
{
  "email": "user@example.com"
}
```

**Success Response (200):**
```json
{
  "message": "If your email is registered, you will receive a password reset link.",
  "data": null
}
```

**Error Responses:**

**400 Bad Request - Invalid Email:**
```json
{
  "message": "Invalid email format"
}
```

### Reset Password

**Endpoint:** `POST /api/v1/auth/reset-password`

**Request Body:**
```json
{
  "token": "reset-token-from-email",
  "password": "newpassword123",
  "password_confirmation": "newpassword123"
}
```

**Success Response (200):**
```json
{
  "message": "Password reset successfully",
  "data": null
}
```

**Error Responses:**

**400 Bad Request - Invalid Token:**
```json
{
  "message": "Invalid or expired reset token"
}
```

**400 Bad Request - Password Validation:**
```json
{
  "message": "Password confirmation does not match"
}
```

**400 Bad Request - Password Too Short:**
```json
{
  "message": "Password must be at least 8 characters long"
}
```

### Get Profile

**Endpoint:** `GET /api/v1/auth/profile`

**Headers:**
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**Success Response (200):**
```json
{
  "message": "Profile retrieved successfully",
  "data": {
    "id": 1,
    "email": "user@example.com",
    "role": "Customer",
    "name": "John Doe",
    "phone": "+628123456789",
    "address": "Jl. Example No. 123",
    "lat": "-6.2088",
    "lng": "106.8456",
    "photo": "https://example.com/photo.jpg"
  }
}
```

**Error Responses:**

**401 Unauthorized - Missing Token:**
```json
{
  "message": "Authorization header required",
  "data": null
}
```

**401 Unauthorized - Invalid Token:**
```json
{
  "message": "Invalid or expired token",
  "data": null
}
```

**404 Not Found - User Not Found:**
```json
{
  "message": "User not found",
  "data": null
}
```

**500 Internal Server Error:**
```json
{
  "message": "Internal server error",
  "data": null
}
```

### Upload Profile Image

**✅ STATUS: SUDAH DI TEST DAN BERFUNGSI**

**UPDATE**: Fitur upload profile image telah **DI TEST** dan **BERFUNGSI DENGAN BAIK**. Fitur sudah siap untuk production dengan automatic cleanup foto lama.

**Fitur**:
- ✅ Upload foto profile ke Supabase Storage
- ✅ Automatic cleanup foto lama saat upload baru
- ✅ Validasi file lengkap (size, type, extension)
- ✅ Error handling yang robust

**Endpoint:** `POST /api/v1/auth/profile/image-upload`

**Headers:**
```
Authorization: Bearer <jwt_token>
Content-Type: multipart/form-data
```

**Form Data:**
- `photo`: File (required) - JPEG, PNG, GIF, WebP, max 5MB

**Success Response (200):**
```json
{
  "message": "Profile image uploaded successfully",
  "data": {
    "image_url": "https://storage.googleapis.com/bucket/profile-uuid.jpg"
  }
}
```

**Error Responses:**

**400 Bad Request - Missing File:**
```json
{
  "message": "Photo is required",
  "data": null
}
```

**401 Unauthorized - Invalid Token:**
```json
{
  "message": "Invalid or expired token",
  "data": null
}
```

**422 Unprocessable Entity - File Too Large:**
```json
{
  "message": "File size too large, maximum 5MB",
  "data": null
}
```

**422 Unprocessable Entity - Invalid File Type:**
```json
{
  "message": "Invalid file type, only JPEG, PNG, GIF, and WebP are allowed",
  "data": null
}
```

**500 Internal Server Error - Upload Failed:**
```json
{
  "message": "Failed to upload image to storage",
  "data": null
}
```

**500 Internal Server Error - Database Update Failed:**
```json
{
  "message": "Failed to update profile",
  "data": null
}
```

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

#### 1. Sign Up - Create User Account
```bash
curl -X POST http://localhost:8080/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newuser@example.com",
    "name": "New User",
    "password": "password123",
    "password_confirmation": "password123"
  }'
```

**Response (201):**
```json
{
  "message": "Account created successfully. Please check your email for verification.",
  "data": {
    "id": 1,
    "name": "New User",
    "email": "newuser@example.com"
  }
}
```

#### 2. Sign In - Success
```bash
curl -X POST http://localhost:8080/api/v1/auth/signin \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

**Response (200):**
```json
{
  "message": "Sign in successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "role": "user",
    "id": 2,
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "+628123456789",
    "lat": "-6.2088",
    "lng": "106.8456"
  }
}
```

#### 2. Admin Check - dengan JWT Token
```bash
curl -X GET http://localhost:8080/api/v1/admin/check \
  -H "Authorization: Bearer YOUR_JWT_TOKEN_HERE"
```

**Response (200):**
```json
{
  "message": "Authentication successful",
  "data": {
    "user_id": 2,
    "email": "john@example.com",
    "role": "user",
    "session_id": "sess_1760512487974112400"
  }
}
```

#### 3. Admin Check - tanpa Token
```bash
curl -X GET http://localhost:8080/api/v1/admin/check
```

**Response (401):**
```json
{
  "message": "Authorization header required",
  "data": null
}
```

**Response (401) - Invalid Token:**
```json
{
  "message": "Invalid or expired token",
  "data": null
}
```

**Response (401) - Session Expired:**
```json
{
  "message": "Session expired or invalid",
  "data": null
}
```

#### 4. Sign In - Invalid Email
```bash
curl -X POST http://localhost:8080/api/v1/auth/signin \
  -H "Content-Type: application/json" \
  -d '{
    "email": "",
    "password": "password123"
  }'
```

#### 5. Sign In - User Not Found
```bash
curl -X POST http://localhost:8080/api/v1/auth/signin \
  -H "Content-Type: application/json" \
  -d '{
    "email": "notfound@example.com",
    "password": "password123"
  }'
```

#### 6. Sign In - Wrong Password
```bash
curl -X POST http://localhost:8080/api/v1/auth/signin \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "wrongpassword"
  }'
```

### Cek JWT Token di Redis

Setelah berhasil sign in, cek apakah token tersimpan di Redis:

#### 1. Cek jumlah keys di Redis
```bash
redis-cli dbsize
# Output: (integer) 2
```

#### 2. Scan session keys
```bash
redis-cli --scan --pattern "session:*"
# Output: "session:2:sess_1760512487974112400"
```

#### 3. Lihat JWT token
```bash
redis-cli get "session:2:sess_1760512487974112400"
# Output: JWT token lengkap
```

#### 4. Lihat session info
```bash
redis-cli hgetall "user_sessions:2"
# Output: JSON dengan session details
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
   - ✅ Admin check with token → Should return 200 with user data
   - ❌ Admin check without token → Should return 401 with `"data": null`

### Using JWT Token

Setelah mendapat token dari Sign In, gunakan untuk API yang protected:

```bash
curl -X GET http://localhost:8080/api/v1/admin/check \
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
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=sayur_user
DATABASE_PASSWORD=sayur_password
DATABASE_NAME=sayur_db
DATABASE_MAX_OPEN_CONNECTION=10
DATABASE_MAX_IDLE_CONNECTION=20

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# RabbitMQ Configuration
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USER=sayur_user
RABBITMQ_PASSWORD=sayur_password
RABBITMQ_VHOST=/
```

## 📦 Dependencies

- **Echo**: Web framework
- **GORM**: ORM untuk database
- **PostgreSQL**: Database
- **JWT**: Token authentication
- **Zerolog**: Structured logging
- **Testify**: Testing framework

## 🐳 Docker

### Docker Compose (Recommended)

```bash
# Start all services (PostgreSQL, Redis, RabbitMQ)
docker-compose up -d

# Check running containers
docker-compose ps

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### Manual Docker Commands

```bash
# Build image
docker build -t user-service .

# Run container
docker run -p 8080:8080 --env-file .env user-service
```

### Monitoring Services

#### RabbitMQ Management UI
- **URL:** http://localhost:15672
- **Username:** sayur_user
- **Password:** sayur_password
- **Check queues:** email_queue

#### Database Admin (Adminer)
- **URL:** http://localhost:8081
- **System:** PostgreSQL
- **Server:** postgres (atau localhost jika local)
- **Username:** sayur_user
- **Password:** sayur_password
- **Database:** sayur_db

#### Redis CLI
```bash
# Connect to Redis
docker exec -it sayur-redis redis-cli

# Check keys
KEYS *

# Check sessions
SCAN 0 MATCH session:*
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
