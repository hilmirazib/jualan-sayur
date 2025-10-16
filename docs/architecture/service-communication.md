# Service Communication Patterns

## Overview
MICRO-SAYUR menggunakan kombinasi synchronous dan asynchronous communication patterns untuk komunikasi antar microservices.

## Communication Patterns

### 1. Synchronous Communication (REST APIs)

#### User Service ↔ Notification Service
```go
// User Service publishes event
userService.Register(user) // → Publish "user.registered" event

// Notification Service consumes event
notificationService.SendWelcomeEmail(user)
```

#### Characteristics
- **Protocol**: HTTP/REST
- **Pattern**: Request-Response
- **Coupling**: Tight coupling
- **Use Case**: Real-time operations, data consistency

#### Implementation
```go
// HTTP Client in User Service
type EmailPublisher interface {
    PublishUserRegistered(ctx context.Context, user *model.User) error
}

// REST API Call
func (p *emailPublisher) PublishUserRegistered(ctx context.Context, user *model.User) error {
    payload := map[string]interface{}{
        "user_id": user.ID,
        "email":   user.Email,
        "name":    user.Name,
    }

    return p.httpClient.Post("/api/notifications/user-registered", payload)
}
```

### 2. Asynchronous Communication (Message Queue)

#### Event-Driven Architecture
```go
// Producer (User Service)
events.Publish("user.registered", userData)

// Consumer (Notification Service)
events.Subscribe("user.registered", func(data UserData) {
    sendWelcomeEmail(data)
})
```

#### Characteristics
- **Protocol**: AMQP (RabbitMQ)
- **Pattern**: Publish-Subscribe
- **Coupling**: Loose coupling
- **Use Case**: Background processing, decoupling services

#### Implementation
```go
// Message Publisher
type MessagePublisher interface {
    Publish(event string, data interface{}) error
}

// RabbitMQ Implementation
func (p *rabbitPublisher) Publish(event string, data interface{}) error {
    body, _ := json.Marshal(data)
    return p.channel.Publish(
        "micro-sayur", // exchange
        event,         // routing key
        false,         // mandatory
        false,         // immediate
        amqp.Publishing{
            ContentType: "application/json",
            Body:        body,
        },
    )
}
```

## Current Implementation

### Active Communications

#### 1. User Registration Flow
```
User Service → RabbitMQ → Notification Service
    ↓              ↓              ↓
Register User → Publish Event → Send Email
```

#### 2. Email Verification Flow
```
User Service → RabbitMQ → Notification Service
    ↓              ↓              ↓
Create Token → Publish Event → Send Verification Email
```

### Message Contracts

#### User Registered Event
```json
{
  "event": "user.registered",
  "timestamp": "2025-10-16T13:52:34Z",
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
  "timestamp": "2025-10-16T13:52:34Z",
  "data": {
    "user_id": "uuid",
    "email": "user@example.com",
    "verification_token": "token",
    "expires_at": "2025-10-16T14:52:34Z"
  }
}
```

## Future Communication Patterns

### Planned Services Communication

#### Order → Payment Service
```
Order Service → REST API → Payment Service
    ↓                      ↓
Create Order → Process Payment → Return Result
```

#### Product → Order Service
```
Product Service → Message Queue → Order Service
    ↓                          ↓
Stock Update → Publish Event → Update Order Status
```

#### API Gateway → All Services
```
API Gateway → REST APIs → All Services
     ↓                      ↓
Route Request → Forward → Process Request
```

## Error Handling

### Synchronous Communication
- Circuit Breaker pattern
- Retry mechanisms
- Timeout handling
- Fallback responses

### Asynchronous Communication
- Dead Letter Queue (DLQ)
- Message retry policies
- Idempotency handling
- Monitoring and alerting

## Monitoring and Observability

### Metrics to Track
- Message throughput
- Processing latency
- Error rates
- Queue depth
- Service health

### Tools
- **Prometheus**: Metrics collection
- **Grafana**: Visualization
- **ELK Stack**: Log aggregation
- **Jaeger**: Distributed tracing

## Best Practices

1. **Define Clear Contracts**: Document message schemas and API contracts
2. **Idempotency**: Ensure operations can be safely retried
3. **Monitoring**: Implement comprehensive monitoring
4. **Error Handling**: Plan for failure scenarios
5. **Versioning**: Handle API and message versioning
6. **Security**: Secure inter-service communication
7. **Testing**: Test communication patterns thoroughly
