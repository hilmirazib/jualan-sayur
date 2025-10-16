# Database Schema

## Overview
MICRO-SAYUR menggunakan PostgreSQL sebagai database utama dengan pola shared database untuk microservices yang ada saat ini.

## Current Database Schema

### Tables Overview

#### 1. users
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    address TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

#### 2. roles
```sql
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

#### 3. user_roles
```sql
CREATE TABLE user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, role_id)
);
```

#### 4. verification_users
```sql
CREATE TABLE verification_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    token_type VARCHAR(50) NOT NULL, -- 'email_verification', 'password_reset'
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

#### 5. sessions
```sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(500) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_activity TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

## Indexes

### Performance Indexes
```sql
-- Users table indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_created_at ON users(created_at);

-- User roles indexes
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);

-- Verification tokens indexes
CREATE INDEX idx_verification_users_user_id ON verification_users(user_id);
CREATE INDEX idx_verification_users_token ON verification_users(token);
CREATE INDEX idx_verification_users_expires_at ON verification_users(expires_at);

-- Sessions indexes
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_token ON sessions(token);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
```

## Relationships

### Entity Relationship Diagram
```
users (1) ──── (N) user_roles (N) ──── (1) roles
   │                                        │
   │                                        │
   ↓                                        │
   └── (1) verification_users (N)           │
   │                                        │
   └── (1) sessions (N)                     │
```

## Data Constraints

### Business Rules
1. **Email Uniqueness**: Satu email hanya bisa digunakan satu user
2. **Role Assignment**: User bisa memiliki multiple roles
3. **Token Validity**: Verification tokens memiliki expiration time
4. **Session Management**: Sessions expired otomatis
5. **Soft Deletes**: Menggunakan flag `is_used` untuk tokens

## Future Schema Extensions

### Planned Tables for Product Service
```sql
-- Products
CREATE TABLE products (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    category_id UUID REFERENCES categories(id),
    stock_quantity INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Categories
CREATE TABLE categories (
    id UUID PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    parent_id UUID REFERENCES categories(id)
);
```

### Planned Tables for Order Service
```sql
-- Orders
CREATE TABLE orders (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    total_amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Order items
CREATE TABLE order_items (
    id UUID PRIMARY KEY,
    order_id UUID REFERENCES orders(id),
    product_id UUID REFERENCES products(id),
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL
);
```

### Planned Tables for Payment Service
```sql
-- Payments
CREATE TABLE payments (
    id UUID PRIMARY KEY,
    order_id UUID REFERENCES orders(id),
    amount DECIMAL(10,2) NOT NULL,
    payment_method VARCHAR(50),
    status VARCHAR(50) DEFAULT 'pending',
    transaction_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

## Database Design Patterns

### 1. Shared Database Pattern
- Saat ini menggunakan satu database untuk semua services
- Akan migrate ke database per service di masa depan

### 2. UUID Primary Keys
- Menggunakan UUID untuk distributed systems compatibility
- Better security (tidak predictable)

### 3. Timestamp Management
- `created_at` dan `updated_at` untuk audit trails
- Automatic timestamps dengan triggers

### 4. Soft Deletes
- Menggunakan flags bukan hard delete
- Maintain data integrity dan audit trails

## Migration Strategy

### Current Migration Files
- `000001_create_users_table.sql`
- `000002_create_roles_table.sql`
- `000003_create_user_role_table.sql`
- `000004_create_verification_users_table.sql`

### Future Migration Planning
1. **Database per Service**: Separate databases untuk setiap service
2. **Event Sourcing**: Untuk complex business logic
3. **CQRS**: Command Query Responsibility Segregation
4. **Read Replicas**: Untuk performance optimization

## Backup and Recovery

### Backup Strategy
- Daily full backups
- Hourly incremental backups
- Point-in-time recovery capability

### High Availability
- PostgreSQL streaming replication
- Automatic failover
- Multi-region deployment (future)

## Performance Considerations

### Query Optimization
- Proper indexing strategy
- Query performance monitoring
- Connection pooling

### Monitoring
- Database performance metrics
- Slow query logs
- Connection pool monitoring
