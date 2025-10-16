# Notification Service

## Overview
Notification Service adalah microservice yang menangani semua operasi terkait pengiriman notifikasi dalam platform MICRO-SAYUR. Service ini bertanggung jawab atas pengiriman email notifikasi secara asynchronous.

## Features

### ✅ Implemented Features
- **Email Notifications**: Pengiriman email untuk berbagai event
- **Asynchronous Processing**: Message queue untuk decoupling
- **Template-based Emails**: Email templates untuk berbagai jenis notifikasi
- **Event-driven Architecture**: Subscribe ke event dari service lain

### 🚧 Planned Features
- SMS notifications
- Push notifications
- Notification preferences
- Email analytics
- Bulk email campaigns
- Notification history

## Architecture

### Clean Architecture (Hexagonal)
```
internal/
├── core/
│   ├── service/         # Email service logic
│   └── port/            # Email service interface
└── adapter/
    ├── consumer/        # Message queue consumer
    └── email/           # Email provider adapter
```

## Supported Email Types

### User Registration
- **Trigger**: User berhasil register
- **Content**: Welcome email dengan verification link
- **Template**: `user_welcome.html`

### Email Verification
- **Trigger**: User meminta verifikasi email
- **Content**: Email dengan verification token
- **Template**: `email_verification.html`

### Password Reset
- **Trigger**: User meminta forgot password
- **Content**: Email dengan reset password link
- **Template**: `password_reset.html`

### Order Confirmations (Future)
- **Trigger**: Order berhasil dibuat
- **Content**: Order details dan invoice
- **Template**: `order_confirmation.html`

## Message Queue Integration

### RabbitMQ Configuration
```go
// Exchange: micro-sayur
// Routing Keys:
// - user.registered
// - user.email_verification
// - user.password_reset
// - order.created (future)
```

### Message Contracts

#### User Registered Event
```json
{
  "event": "user.registered",
  "timestamp": "2025-10-16T13:55:39Z",
  "data": {
    "user_id": "uuid",
    "email": "user@example.com",
    "name": "John Doe",
    "verification_token": "token"
  }
}
```

#### Email Verification Event
```json
{
  "event": "user.email_verification",
  "timestamp": "2025-10-16T13:55:39Z",
  "data": {
    "user_id": "uuid",
    "email": "user@example.com",
    "verification_token": "token",
    "expires_at": "2025-10-16T14:55:39Z"
  }
}
```

#### Password Reset Event
```json
{
  "event": "user.password_reset",
  "timestamp": "2025-10-16T13:55:39Z",
  "data": {
    "user_id": "uuid",
    "email": "user@example.com",
    "reset_token": "token",
    "expires_at": "2025-10-16T14:55:39Z"
  }
}
```

## Email Templates

### Template Structure
```
templates/
├── user_welcome.html
├── email_verification.html
├── password_reset.html
└── base.html
```

### Template Variables
```go
type EmailData struct {
    RecipientName string
    RecipientEmail string
    VerificationURL string
    ResetURL string
    CompanyName string
    SupportEmail string
}
```

## Dependencies

### External Services
- **RabbitMQ**: Message queue untuk event consumption
- **SMTP Server**: Email delivery (configurable)

### Internal Dependencies
- Tidak ada direct dependencies ke service lain
- Receives events via message queue

## Configuration

### Environment Variables
```bash
# RabbitMQ
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest
RABBITMQ_VHOST=/

# Email Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM_EMAIL=noreply@micro-sayur.com
SMTP_FROM_NAME=MICRO-SAYUR

# Server
SERVER_PORT=8081
SERVER_HOST=0.0.0.0
```

## Development

### Prerequisites
- Go 1.21+
- RabbitMQ 3.12+

### Setup
```bash
# Clone repository
git clone https://github.com/hilmirazib/jualan-sayur.git
cd services/notification-service

# Install dependencies
go mod download

# Copy environment file
cp .env.example .env

# Run service
make run
```

### Testing
```bash
# Run unit tests
make test

# Run integration tests
make test-integration

# Test email sending
make test-email
```

### Available Commands
```bash
make run              # Start the service
make build            # Build binary
make test             # Run tests
make docker-build     # Build Docker image
make docker-run       # Run with Docker
make test-email       # Test email functionality
```

## Deployment

### Docker
```bash
# Build image
docker build -t micro-sayur/notification-service .

# Run container
docker run -p 8081:8081 micro-sayur/notification-service
```

### Docker Compose (Development)
```yaml
version: '3.8'
services:
  notification-service:
    build: .
    ports:
      - "8081:8081"
    environment:
      - RABBITMQ_HOST=rabbitmq
      - SMTP_HOST=smtp-server
    depends_on:
      - rabbitmq
```

## Email Providers

### Supported Providers
1. **SMTP** (Default) - Direct SMTP connection
2. **SendGrid** - Cloud email service
3. **Mailgun** - Transactional email service
4. **AWS SES** - Amazon Simple Email Service

### Provider Configuration
```go
type EmailProvider interface {
    SendEmail(to, subject, htmlBody string) error
}

type SMTPProvider struct {
    host     string
    port     int
    username string
    password string
}
```

## Monitoring & Health Checks

### Health Endpoints
```
GET  /health     # Overall health status
GET  /ready      # Readiness probe
GET  /metrics    # Prometheus metrics
```

### Key Metrics
- Messages processed per second
- Email delivery success rate
- Queue depth
- Processing latency
- Error rates by email type

## Error Handling

### Message Processing Errors
- **Retry Logic**: Exponential backoff untuk failed messages
- **Dead Letter Queue**: Messages yang gagal dipindahkan ke DLQ
- **Alerting**: Notification untuk persistent failures

### Email Delivery Errors
- **Bounce Handling**: Automatic bounce processing
- **Unsubscribe**: Honor unsubscribe requests
- **Rate Limiting**: Respect provider limits

### Error Response Format
```json
{
  "error": {
    "code": "EMAIL_SEND_FAILED",
    "message": "Failed to send email",
    "details": {
      "recipient": "user@example.com",
      "error": "SMTP connection timeout"
    }
  }
}
```

## Logging

### Log Levels
- `DEBUG` - Detailed debug information
- `INFO` - General information (message processed, email sent)
- `WARN` - Warning messages (retries, temporary failures)
- `ERROR` - Error conditions (permanent failures)

### Structured Logging
```json
{
  "level": "INFO",
  "timestamp": "2025-10-16T13:55:39Z",
  "service": "notification-service",
  "event": "email.sent",
  "recipient": "user@example.com",
  "email_type": "user_welcome",
  "processing_time_ms": 250,
  "status": "success"
}
```

## Security

### Email Security
- **DKIM/SPF**: Email authentication
- **TLS Encryption**: Secure SMTP connections
- **Input Validation**: Sanitize email content
- **Rate Limiting**: Prevent email abuse

### Data Protection
- **PII Handling**: Secure handling of personal data
- **Audit Logging**: Track all email sends
- **Compliance**: GDPR/CCPA compliance

## Performance

### Throughput
- **Target**: 100 emails/second
- **Current**: 50 emails/second
- **Scaling**: Horizontal scaling dengan multiple instances

### Latency
- **Average**: < 500ms per email
- **P95**: < 2 seconds
- **P99**: < 5 seconds

## Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/new-email-type`)
3. Add email template
4. Update message consumer
5. Add tests
6. Create Pull Request

## API Documentation

### Health Check Endpoints
```
GET  /health     # Service health
GET  /ready      # Service readiness
GET  /metrics    # Prometheus metrics
```

### Internal Endpoints (Development)
```
POST /api/v1/test-email    # Send test email
GET  /api/v1/templates     # List available templates
```

## Troubleshooting

### Common Issues

#### RabbitMQ Connection Failed
```bash
# Check RabbitMQ status
docker ps | grep rabbitmq

# Check connection logs
docker logs notification-service
```

#### Email Delivery Failed
```bash
# Check SMTP credentials
# Verify email templates
# Check SMTP server logs
```

#### High Queue Depth
```bash
# Check service logs for errors
# Verify email provider limits
# Scale service instances
```

## Future Enhancements

### Advanced Features
- **Email Analytics**: Open rates, click tracking
- **A/B Testing**: Template optimization
- **Personalization**: Dynamic content
- **Scheduling**: Delayed email delivery
- **Webhooks**: Delivery status callbacks

### Integration Plans
- **CRM Integration**: Customer data sync
- **Marketing Automation**: Campaign management
- **Notification Preferences**: User opt-in/opt-out
