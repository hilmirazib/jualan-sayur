# MICRO-SAYUR Documentation

## Overview
MICRO-SAYUR adalah platform microservices untuk sistem jual beli sayuran yang dibangun dengan arsitektur hexagonal (clean architecture).

## Services

### User Service
- Manajemen pengguna dan autentikasi
- Port: 8080
- Teknologi: Go, PostgreSQL, Redis, RabbitMQ
- Features: Registration, Login, JWT Auth, Email Verification, Password Reset

### Notification Service
- Pengiriman email notifikasi
- Port: 8081
- Teknologi: Go, RabbitMQ
- Features: Email Templates, Async Processing, Multiple Email Types

### Product Service (Planned)
- Manajemen produk sayuran
- Port: 8082

### Order Service (Planned)
- Proses pemesanan
- Port: 8083

### Payment Service (Planned)
- Integrasi pembayaran
- Port: 8084

## Architecture
Lihat folder `architecture/` untuk diagram dan dokumentasi arsitektur sistem.

## API Documentation
Lihat folder `api/` untuk spesifikasi OpenAPI/Swagger setiap service.

## Deployment
Lihat folder `deployment/` untuk panduan deployment dan konfigurasi infrastruktur.

## Development
Lihat README.md di setiap service untuk panduan development lokal.
