# Sayur Project - Microservices E-Commerce Platform

![Project Status](https://img.shields.io/badge/status-in%20development-yellow)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue)
![License](https://img.shields.io/badge/license-MIT-green)

Sayur Project adalah platform e-commerce berbasis microservices yang dibangun dengan **Golang**, **Nuxt.js**, dan **PostgreSQL**. Platform ini berfokus pada penjualan produk segar seperti sayur-mayur, buah-buahan, dan bahan makanan lainnya dengan sistem delivery radius terbatas untuk menjaga kualitas produk.

## 🏗️ Arsitektur Sistem

Platform ini menggunakan **microservices architecture** dengan 5 layanan utama:

- **User Service** - Manajemen user, autentikasi, dan otorisasi
- **Product Service** - Katalog produk, kategori, dan inventory management  
- **Order Service** - Proses pemesanan, shopping cart, dan order tracking
- **Payment Service** - Integrasi payment gateway (Midtrans) dan transaction logging
- **Notification Service** - Email notifications dan push notifications

## 🛠️ Technology Stack

### Backend
- **Language**: Go (Golang) 1.21+
- **Web Framework**: Echo Framework
- **Database**: PostgreSQL 15+
- **Cache**: Redis 7+
- **Message Broker**: RabbitMQ
- **Authentication**: JWT

### Frontend  
- **Framework**: Nuxt.js 3
- **UI Library**: Vue.js 3
- **CSS Framework**: Tailwind CSS

### Infrastructure
- **Containerization**: Docker & Docker Compose
- **Orchestration**: Kubernetes
- **Load Testing**: K6
- **Payment Gateway**: Midtrans

## 🚀 Quick Start

### Prerequisites
- Go 1.21 atau lebih baru
- Node.js 18 atau lebih baru
- PostgreSQL 15+
- Redis 7+
- RabbitMQ 3.12+
- Docker & Docker Compose (optional)

### Local Development Setup

1. **Clone repository**
   ```bash
   git clone <repository-url>
   cd micro-sayur
   ```

2. **Setup environment variables**
   ```bash
   cp .env.example .env
   # Edit .env file dengan konfigurasi database dan services
   ```

3. **Start infrastructure services**
   ```bash
   docker-compose up -d postgres redis rabbitmq
   ```

4. **Run database migrations**
   ```bash
   make migrate-up
   ```

5. **Start services**
   ```bash
   # Terminal 1 - User Service
   cd services/user-service
   go run cmd/main.go
   
   # Terminal 2 - Product Service  
   cd services/product-service
   go run cmd/main.go
   
   # Terminal 3 - Order Service
   cd services/order-service
   go run cmd/main.go
   
   # Terminal 4 - Payment Service
   cd services/payment-service
   go run cmd/main.go
   
   # Terminal 5 - Notification Service
   cd services/notification-service
   go run cmd/main.go
   ```

## 📚 API Documentation

- **User Service**: http://localhost:8001/swagger
- **Product Service**: http://localhost:8002/swagger  
- **Order Service**: http://localhost:8003/swagger
- **Payment Service**: http://localhost:8004/swagger
- **Notification Service**: http://localhost:8005/swagger

## 🏪 Business Logic & Constraints

### Geographic Constraints
- **Delivery Radius**: Maksimal 1-5 KM dari toko untuk menjaga kualitas produk segar
- **Location Validation**: Sistem validasi otomatis berdasarkan koordinat GPS

### Operational Hours  
- **Order Cutoff**: Pemesanan ditutup jam 21:00 - 09:00 setiap hari
- **Fresh Product Focus**: Khusus produk segar dengan manajemen expired date

### Shipping Options
- **Pickup**: Gratis, customer datang ke toko
- **Delivery**: Biaya tambahan Rp 5.000

## 🧪 Testing

```bash
# Run all tests
make test

# Run specific service tests
make test-user-service
make test-product-service  
make test-order-service
make test-payment-service
make test-notification-service

# Run integration tests
make test-integration

# Run load tests
make test-load
```

## 🚢 Deployment

### Using Docker Compose
```bash
# Build dan start semua services
docker-compose up --build

# Production deployment
docker-compose -f docker-compose.prod.yml up -d
```

### Using Kubernetes
```bash
# Deploy ke Kubernetes cluster
kubectl apply -f deployments/k8s/

# Check deployment status
kubectl get pods -n sayur-project
```

## 📖 Documentation

- [System Architecture](docs/architecture.md)
- [API Specifications](docs/api.md)
- [Database Schema](docs/database.md)
- [Deployment Guide](docs/deployment.md)
- [Development Guidelines](docs/development.md)

## 🤝 Contributing

1. Fork repository ini
2. Buat feature branch (`git checkout -b feature/amazing-feature`)
3. Commit perubahan (`git commit -m 'Add some amazing feature'`)
4. Push ke branch (`git push origin feature/amazing-feature`)
5. Buat Pull Request

## 📝 License

Project ini dilisensikan di bawah MIT License. Lihat file [LICENSE](LICENSE) untuk detail lengkap.

## 👥 Team

- **Backend Developer**: [Your Name]
- **Frontend Developer**: [Frontend Dev Name]  
- **DevOps Engineer**: [DevOps Engineer Name]

## 📞 Support

Jika ada pertanyaan atau butuh bantuan, silakan:
- Buat [GitHub Issue](../../issues)
- Contact: [your-email@example.com]

---

**Made with ❤️ for Indonesian fresh food market**