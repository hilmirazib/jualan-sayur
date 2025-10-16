# System Overview

## MICRO-SAYUR Platform

MICRO-SAYUR adalah platform e-commerce khusus untuk jual beli sayuran yang dibangun dengan arsitektur microservices modern.

## Business Domain

### Core Business Processes
1. **User Management**
   - Registrasi dan login user
   - Manajemen profil
   - Role-based access control

2. **Product Management**
   - Katalog produk sayuran
   - Manajemen inventory
   - Kategori dan sub-kategori

3. **Order Management**
   - Proses pemesanan
   - Shopping cart
   - Order tracking

4. **Payment Processing**
   - Integrasi payment gateway
   - Payment verification
   - Refund handling

5. **Notification System**
   - Email notifications
   - Order confirmations
   - Status updates

## System Goals

### Functional Requirements
- User registration and authentication
- Product browsing and search
- Shopping cart management
- Order placement and tracking
- Payment processing
- Email notifications

### Non-Functional Requirements
- High availability (99.9% uptime)
- Scalability (handle 1000+ concurrent users)
- Security (data encryption, secure APIs)
- Performance (response time < 500ms)
- Maintainability (clean architecture)

## Current Architecture State

### Active Services
- **User Service**: ✅ Implemented
- **Notification Service**: ✅ Implemented

### Planned Services
- **Product Service**: 📋 Planned
- **Order Service**: 📋 Planned
- **Payment Service**: 📋 Planned
- **API Gateway**: 📋 Planned

### Infrastructure
- **Database**: PostgreSQL (shared database pattern)
- **Cache**: Redis
- **Message Queue**: RabbitMQ
- **Containerization**: Docker
- **Orchestration**: Docker Compose

## Technology Decisions

### Programming Language
- **Go**: Chosen for performance, concurrency, and microservices suitability

### Architecture Pattern
- **Hexagonal Architecture**: For testability and maintainability
- **Domain-Driven Design**: For complex business logic

### Communication
- **REST APIs**: For synchronous communication
- **Message Queue**: For asynchronous event-driven communication

### Deployment
- **Docker**: For containerization
- **Docker Compose**: For local development
- **Kubernetes**: For production deployment (planned)
