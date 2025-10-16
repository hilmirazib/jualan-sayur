# Shared Packages

Folder ini berisi package-package yang dapat digunakan bersama oleh semua microservices dalam sistem MICRO-SAYUR.

## Struktur

### models/
Berisi model-model data yang shared antar services, seperti:
- Common response models
- Error models
- Pagination models
- Base entity models

### utils/
Berisi utility functions yang reusable, seperti:
- String manipulation
- Date/time helpers
- Validation helpers
- Logging helpers

### auth/
Berisi komponen autentikasi dan otorisasi yang shared, seperti:
- JWT token helpers
- Role/permission models
- Authentication middleware

### database/
Berisi helpers untuk koneksi dan operasi database yang shared, seperti:
- Connection pooling
- Migration helpers
- Query builders
- Transaction helpers

## Usage

Untuk menggunakan package ini dari service lain, import dengan path relatif: (coming soon)

```go
import "github.com/hilmirazib/jualan-sayur/pkg/models"
import "github.com/hilmirazib/jualan-sayur/pkg/utils"
```

## Development Guidelines

1. Pastikan semua package backward compatible
2. Gunakan interface untuk dependency injection
3. Sertakan unit tests untuk setiap package
4. Dokumentasikan fungsi dan struct dengan komentar Go
5. Ikuti clean architecture principles
