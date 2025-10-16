# Deployment Architecture

## Overview
MICRO-SAYUR menggunakan containerization dengan Docker dan orchestration dengan Docker Compose untuk development, dengan rencana migrasi ke Kubernetes untuk production.

## Current Deployment Architecture

### Development Environment

#### Docker Compose Setup
```yaml
version: '3.8'
services:
  user-service:
    build: ./services/user-service
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
      - RABBITMQ_HOST=rabbitmq
    depends_on:
      - postgres
      - redis
      - rabbitmq

  notification-service:
    build: ./services/notification-service
    ports:
      - "8081:8081"
    environment:
      - RABBITMQ_HOST=rabbitmq
    depends_on:
      - rabbitmq

  postgres:
    image: postgres:15
    environment:
      - POSTGRES_DB=micro_sayur
      - POSTGRES_USER=micro_sayur
      - POSTGRES_PASSWORD=password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  rabbitmq:
    image: rabbitmq:3-management
    ports:
      - "5672:5672"
      - "15672:15672"
    environment:
      - RABBITMQ_DEFAULT_USER=guest
      - RABBITMQ_DEFAULT_PASS=guest

volumes:
  postgres_data:
```

### Infrastructure Components

#### 1. Application Services
- **User Service**: Port 8080, Go application
- **Notification Service**: Port 8081, Go application

#### 2. Data Stores
- **PostgreSQL**: Primary database, port 5432
- **Redis**: Cache dan session store, port 6379

#### 3. Message Queue
- **RabbitMQ**: Event-driven communication, ports 5672 (AMQP), 15672 (Management UI)

## Production Deployment Strategy

### Kubernetes Migration Plan

#### 1. Container Registry
```yaml
# Dockerfile untuk setiap service
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
```

#### 2. Kubernetes Manifests Structure
```
deployments/
├── k8s/
│   ├── base/
│   │   ├── user-service/
│   │   │   ├── deployment.yaml
│   │   │   ├── service.yaml
│   │   │   ├── configmap.yaml
│   │   │   └── hpa.yaml
│   │   ├── notification-service/
│   │   └── infrastructure/
│   │       ├── postgres/
│   │       ├── redis/
│   │       └── rabbitmq/
│   └── overlays/
│       ├── development/
│       ├── staging/
│       └── production/
```

#### 3. Service Deployment
```yaml
# user-service deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: user-service
  template:
    metadata:
      labels:
        app: user-service
    spec:
      containers:
      - name: user-service
        image: micro-sayur/user-service:v1.0.0
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          valueFrom:
            configMapKeyRef:
              name: app-config
              key: db.host
        - name: REDIS_HOST
          valueFrom:
            configMapKeyRef:
              name: app-config
              key: redis.host
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

#### 4. Service Exposure
```yaml
# user-service service.yaml
apiVersion: v1
kind: Service
metadata:
  name: user-service
spec:
  selector:
    app: user-service
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP
```

#### 5. Ingress Configuration
```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: micro-sayur-ingress
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
  - host: api.micro-sayur.com
    http:
      paths:
      - path: /users
        pathType: Prefix
        backend:
          service:
            name: user-service
            port:
              number: 80
      - path: /notifications
        pathType: Prefix
        backend:
          service:
            name: notification-service
            port:
              number: 80
```

### Infrastructure as Code

#### Terraform Configuration
```hcl
# main.tf
terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
  }
}

provider "kubernetes" {
  config_path = "~/.kube/config"
}

# VPC and Networking
resource "aws_vpc" "micro_sayur" {
  cidr_block = "10.0.0.0/16"
}

# EKS Cluster
resource "aws_eks_cluster" "micro_sayur" {
  name     = "micro-sayur-cluster"
  role_arn = aws_iam_role.eks_cluster.arn

  vpc_config {
    subnet_ids = aws_subnet.micro_sayur[*].id
  }
}

# RDS PostgreSQL
resource "aws_db_instance" "postgres" {
  allocated_storage    = 20
  engine              = "postgres"
  engine_version      = "15.3"
  instance_class      = "db.t3.micro"
  db_name             = "micro_sayur"
  username            = "micro_sayur"
  password            = var.db_password
  parameter_group_name = "default.postgres15"
  skip_final_snapshot = true
}

# ElastiCache Redis
resource "aws_elasticache_cluster" "redis" {
  cluster_id      = "micro-sayur-redis"
  engine          = "redis"
  node_type       = "cache.t3.micro"
  num_cache_nodes = 1
  port            = 6379
}

# MSK (Managed Streaming for Kafka) or RabbitMQ
resource "aws_msk_cluster" "rabbitmq" {
  cluster_name           = "micro-sayur-messaging"
  kafka_version         = "2.8.1"
  number_of_broker_nodes = 3
}
```

## CI/CD Pipeline

### GitHub Actions Workflow
```yaml
# .github/workflows/deploy.yml
name: Deploy to Kubernetes

on:
  push:
    branches: [ main ]

jobs:
  build-and-deploy:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v3

    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v2

    - name: Login to Docker Hub
      uses: docker/login-action@v2
      with:
        username: ${{ secrets.DOCKER_USERNAME }}
        password: ${{ secrets.DOCKER_PASSWORD }}

    - name: Build and push User Service
      uses: docker/build-push-action@v4
      with:
        context: ./services/user-service
        push: true
        tags: micro-sayur/user-service:latest

    - name: Build and push Notification Service
      uses: docker/build-push-action@v4
      with:
        context: ./services/notification-service
        push: true
        tags: micro-sayur/notification-service:latest

    - name: Deploy to Kubernetes
      uses: azure/k8s-deploy@v4
      with:
        namespace: micro-sayur
        manifests: |
          deployments/k8s/base/user-service/deployment.yaml
          deployments/k8s/base/user-service/service.yaml
          deployments/k8s/base/notification-service/deployment.yaml
          deployments/k8s/base/notification-service/service.yaml
```

## Monitoring and Observability

### Prometheus & Grafana
```yaml
# prometheus.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
data:
  prometheus.yml: |
    global:
      scrape_interval: 15s
    scrape_configs:
    - job_name: 'user-service'
      static_configs:
      - targets: ['user-service:8080']
    - job_name: 'notification-service'
      static_configs:
      - targets: ['notification-service:8081']
```

### Logging
- **EFK Stack**: Elasticsearch, Fluentd, Kibana
- **Centralized Logging**: Semua service logs dikumpulkan
- **Log Aggregation**: Structured logging dengan correlation IDs

### Health Checks
```go
// Health check endpoint
func (h *handler) HealthCheck(c *gin.Context) {
    health := &HealthStatus{
        Status:    "healthy",
        Timestamp: time.Now(),
        Services: map[string]string{
            "database": h.checkDatabase(),
            "redis":    h.checkRedis(),
            "rabbitmq": h.checkRabbitMQ(),
        },
    }
    c.JSON(200, health)
}
```

## Security Considerations

### Network Security
- **Network Policies**: Isolation antar services
- **Service Mesh**: Istio untuk mTLS dan traffic management
- **API Gateway**: Authentication dan rate limiting

### Secret Management
- **Kubernetes Secrets**: Untuk database credentials
- **AWS Secrets Manager**: Untuk production secrets
- **Environment Variables**: Tidak untuk sensitive data

### Backup and Disaster Recovery
- **Database Backups**: Automated daily backups
- **Multi-region**: Cross-region replication
- **Disaster Recovery**: RTO < 4 hours, RPO < 1 hour

## Scaling Strategy

### Horizontal Pod Autoscaling
```yaml
# hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: user-service-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: user-service
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

### Database Scaling
- **Read Replicas**: Untuk read-heavy workloads
- **Connection Pooling**: PgBouncer untuk PostgreSQL
- **Sharding**: Future consideration untuk massive scale

This deployment architecture provides a solid foundation for MICRO-SAYUR's growth from development to production scale.
