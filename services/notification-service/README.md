# Notification Service

Notification service untuk mengirim email menggunakan RabbitMQ sebagai message queue dan Mailtrap sebagai SMTP server untuk testing.

## Features

- RabbitMQ Consumer untuk email queue
- SMTP Email sending dengan gomail
- Mailtrap integration untuk testing
- Graceful shutdown
- Docker support

## Environment Variables

```env
APP_ENV="development"

RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USER=sayur_user
RABBITMQ_PASSWORD=sayur_password
RABBITMQ_VHOST=/

SMTP_HOST=smtp.mailtrap.io
SMTP_PORT=2525
SMTP_USER=your_mailtrap_username
SMTP_PASSWORD=your_mailtrap_password
```

## Setup Mailtrap

1. Daftar di [Mailtrap](https://mailtrap.io)
2. Buat inbox baru
3. Copy SMTP credentials dari Settings > SMTP Settings
4. Update `SMTP_USER` dan `SMTP_PASSWORD` di `.env`

## Running

### Prerequisites
1. **Setup Mailtrap Credentials**
   - Daftar di [Mailtrap](https://mailtrap.io)
   - Buat inbox baru
   - Copy SMTP credentials dari Settings → SMTP Settings
   - Update `.env` file:
   ```env
   SMTP_USER=your_mailtrap_username
   SMTP_PASSWORD=your_mailtrap_password
   ```

2. **Pastikan Dependencies Running**
   ```bash
   # Jalankan RabbitMQ dan database
   docker-compose up postgres redis rabbitmq -d
   ```

### Local Development
```bash
cd services/notification-service
go run cmd/server/main.go
```

### Docker Compose (Production)
```bash
# Jalankan semua services termasuk notification-service
docker-compose up

# Atau jalankan notification-service saja
docker-compose up notification-service
```

### Build & Run Binary
```bash
cd services/notification-service
go build -o notification-service cmd/server/main.go
./notification-service
```

### Docker Build Manual
```bash
cd services/notification-service
docker build -t notification-service .
docker run --env-file .env notification-service
```

## Architecture

```
User Service → RabbitMQ Queue → Notification Service → SMTP (Mailtrap) → Email
```

1. User service publish email message ke RabbitMQ queue
2. Notification service consume message dari queue
3. Kirim email menggunakan SMTP ke Mailtrap
4. Email dapat dilihat di Mailtrap inbox untuk testing
