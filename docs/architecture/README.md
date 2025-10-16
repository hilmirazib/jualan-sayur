# Arsitektur Sistem MICRO-SAYUR

## Overview
MICRO-SAYUR mengadopsi arsitektur microservices dengan clean architecture (hexagonal architecture) untuk memastikan skalabilitas, maintainability, dan testability yang tinggi.

## Struktur Arsitektur

### 1. Microservices Architecture
- **User Service**: Manajemen user, autentikasi, dan autorisasi
- **Notification Service**: Pengiriman email dan notifikasi
- **Product Service**: Manajemen produk sayuran (planned)
- **Order Service**: Proses pemesanan (planned)
- **Payment Service**: Integrasi pembayaran (planned)

### 2. Clean Architecture (Hexagonal)
Setiap service mengikuti prinsip clean architecture dengan layer:
- **Domain Layer**: Business logic dan entities
- **Application Layer**: Use cases dan application services
- **Infrastructure Layer**: External concerns (database, messaging, web)

### 3. Technology Stack
- **Language**: Go
- **Database**: PostgreSQL
- **Cache**: Redis
- **Message Queue**: RabbitMQ
- **API Gateway**: (planned)
- **Container**: Docker
- **Orchestration**: Docker Compose / Kubernetes

## Communication Patterns
- **Synchronous**: REST API calls antar services
- **Asynchronous**: Message queue untuk event-driven communication
- **Database**: Shared database pattern (sementara)

## Files in This Directory
- `system-overview.md` - Gambaran keseluruhan sistem
- `hexagonal-architecture.md` - Penjelasan clean architecture
- `service-communication.md` - Pola komunikasi antar service
- `database-schema.md` - Skema database
- `deployment-architecture.md` - Arsitektur deployment
