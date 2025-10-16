# Hexagonal Architecture (Clean Architecture)

## Overview
MICRO-SAYUR mengimplementasikan Hexagonal Architecture (juga dikenal sebagai Clean Architecture) untuk memastikan separation of concerns, testability, dan maintainability yang tinggi.

## Architecture Layers

### 1. Domain Layer (Core Business Logic)
**Location**: `internal/core/domain/`
**Responsibilities**:
- Business entities dan value objects
- Business rules dan invariants
- Domain models dan DTOs

**Components**:
- `entity/`: Domain entities (User, Role, etc.)
- `model/`: Domain models dan DTOs

### 2. Application Layer (Use Cases)
**Location**: `internal/core/service/`
**Responsibilities**:
- Application use cases
- Orchestration of domain objects
- Transaction management

**Components**:
- Service interfaces (ports)
- Service implementations (use cases)

### 3. Infrastructure Layer (External Concerns)
**Location**: `internal/adapter/`
**Responsibilities**:
- External system integrations
- Data persistence
- External API calls
- Framework-specific code

**Components**:
- `handler/`: HTTP handlers, message consumers
- `repository/`: Data access implementations
- `middleware/`: Cross-cutting concerns
- `message/`: Message publishers/consumers

## Dependency Inversion Principle

### Ports and Adapters Pattern
- **Ports**: Interfaces yang mendefinisikan kontrak (dalam `internal/core/port/`)
- **Adapters**: Implementasi konkrit dari ports (dalam `internal/adapter/`)

### Dependency Direction
```
Infrastructure Layer → Application Layer → Domain Layer
```

Domain layer tidak bergantung pada layer lain, tetapi layer lain bergantung pada domain.

## Benefits

### Testability
- Domain logic dapat ditest tanpa dependencies eksternal
- Mock adapters untuk testing
- Unit tests yang isolated

### Maintainability
- Changes di infrastructure tidak mempengaruhi domain logic
- Framework dapat diganti tanpa mengubah business logic
- Clear separation of concerns

### Flexibility
- Multiple implementations untuk satu port
- Easy to add new features
- Technology-agnostic domain logic

## Implementation in MICRO-SAYUR

### Example: User Service

```
internal/core/
├── domain/
│   ├── entity/user_entity.go
│   └── model/user_model.go
├── port/
│   ├── user_repository_port.go
│   └── user_service_port.go
└── service/
    └── user_service.go

internal/adapter/
├── handler/user_handler.go
├── repository/user_repository.go
└── middleware/jwt_middleware.go
```

### Dependency Injection
```go
// Port (Interface)
type UserRepository interface {
    Create(ctx context.Context, user *entity.User) error
    FindByEmail(ctx context.Context, email string) (*entity.User, error)
}

// Adapter (Implementation)
type userRepositoryImpl struct {
    db *sql.DB
}

func (r *userRepositoryImpl) Create(ctx context.Context, user *entity.User) error {
    // Implementation
}
```

## Testing Strategy

### Unit Tests
- Test domain entities dan business rules
- Mock external dependencies
- Test service layer dengan mock repositories

### Integration Tests
- Test adapter implementations
- Test database operations
- Test external API integrations

### E2E Tests
- Test complete user journeys
- Test service interactions
- Test through API endpoints

## Best Practices

1. **Domain First**: Always start with domain modeling
2. **Dependency Injection**: Use interfaces for all external dependencies
3. **Single Responsibility**: Each layer has clear responsibilities
4. **Test Coverage**: Maintain high test coverage, especially for domain logic
5. **Documentation**: Document ports and their contracts clearly
