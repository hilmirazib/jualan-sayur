# API Testing Guide - MICRO-SAYUR

## 🧪 Hoppscotch Web Setup

Project ini menggunakan **Hoppscotch Web** (https://hoppscotch.io) untuk API testing - tidak perlu setup Docker!

### 🚀 **Quick Setup**

1. **Start API Services:**
```bash
docker-compose up -d
make migrate-up
```

2. **Import Collections:**
   - Buka https://hoppscotch.io
   - Klik "Import" → "Import from JSON"
   - Upload file: `scripts/hoppscotch-web-collections.json`

3. **Start Testing!** ✅

### 📁 **Available Collections**

Setelah import, Anda akan memiliki **11 API endpoints** siap pakai:

#### 🔓 **Public Endpoints:**
- `POST /auth/signup` - Register user
- `POST /auth/signin` - Login user
- `GET /auth/verify` - Verify email
- `POST /auth/forgot-password` - Forgot password
- `POST /auth/reset-password` - Reset password

#### 🔐 **Protected Endpoints:**
- `POST /auth/logout` - Logout user
- `GET /auth/profile` - Get profile
- `POST /auth/profile/image-upload` - **Upload foto profile** ⭐
- `GET /admin/check` - Admin check

#### 🏥 **Health Check:**
- `GET /health` - Service health
- `GET /` - API info

### 🔑 **Environment Variables**

Collections sudah include variables:
- `{{BASE_URL}}` → `http://localhost:8001/api/v1`
- `{{JWT_TOKEN}}` → Update dengan token dari login response

## 📋 Testing Scenarios

### 1. User Registration & Login

#### Register User Baru
```
Method: POST
URL: http://localhost:8001/api/v1/auth/signup
Content-Type: application/json

Body:
{
  "name": "Test User",
  "email": "test@example.com",
  "password": "password123",
  "phone": "081234567890"
}
```

#### Login User
```
Method: POST
URL: http://localhost:8001/api/v1/auth/signup
Content-Type: application/json

Body:
{
  "email": "test@example.com",
  "password": "password123"
}
```

**Expected Response:**
```json
{
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 3600
  }
}
```

### 2. Profile Image Upload

#### Setup Environment Variable
Simpan JWT token dari login response ke environment variable di Hoppscotch:
- Name: `JWT_TOKEN`
- Value: `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...`

#### Upload Foto Profile
```
Method: POST
URL: http://localhost:8001/api/v1/auth/profile/image-upload
Authorization: {{JWT_TOKEN}}
Content-Type: multipart/form-data

Body (Form Data):
- Key: photo
- Type: File
- Value: [Pilih file gambar JPG/PNG max 5MB]
```

**Expected Response:**
```json
{
  "message": "Profile image uploaded successfully",
  "data": {
    "image_url": "https://your-project-id.supabase.co/storage/v1/object/public/profile-images/profile-uuid.jpg"
  }
}
```

### 3. Get User Profile

#### Get Profile Data
```
Method: GET
URL: http://localhost:8001/api/v1/auth/profile
Authorization: {{JWT_TOKEN}}
```

**Expected Response:**
```json
{
  "message": "Profile retrieved successfully",
  "data": {
    "id": 1,
    "name": "Test User",
    "email": "test@example.com",
    "phone": "081234567890",
    "photo": "https://your-project-id.supabase.co/storage/v1/object/public/profile-images/profile-uuid.jpg",
    "created_at": "2025-01-20T10:00:00Z",
    "updated_at": "2025-01-20T10:30:00Z"
  }
}
```

## 🔧 Hoppscotch Features

### Environment Variables
- **Global Variables**: Untuk menyimpan JWT token
- **Environment Switching**: Development/Production environments

### Request Collections
- **Organize Requests**: Group by feature (Auth, Profile, etc.)
- **Save & Reuse**: Tidak perlu setup ulang
- **Import/Export**: Bisa backup collections

### Testing Features
- **Response Validation**: Test response status, schema
- **Environment Variables**: Dynamic value injection
- **Request History**: Track semua requests

## 🐳 Docker Commands

```bash
# Start Hoppscotch
docker-compose up -d hoppscotch

# Stop Hoppscotch
docker-compose stop hoppscotch

# Restart Hoppscotch
docker-compose restart hoppscotch

# View logs
docker-compose logs hoppscotch

# Remove container
docker-compose down hoppscotch
```

## 🔍 Troubleshooting

### Hoppscotch tidak bisa diakses
```bash
# Check container status
docker-compose ps hoppscotch

# Check logs
docker-compose logs hoppscotch

# Restart service
docker-compose restart hoppscotch
```

### Port 3000 conflict
Jika port 3000 sudah digunakan, edit `docker-compose.yml`:
```yaml
hoppscotch:
  ports:
    - "3001:3000"  # Change to available port
```

### API Testing Tips

1. **Save JWT Token**: Setelah login, simpan token ke environment variable
2. **Test File Upload**: Pastikan file gambar < 5MB dan format JPG/PNG/GIF/WebP
3. **Check Image URL**: URL yang dikembalikan harus accessible di browser
4. **Error Handling**: Test dengan invalid token, file terlalu besar, dll.

## 📚 Additional Resources

- [Hoppscotch Documentation](https://docs.hoppscotch.io/)
- [Supabase Storage Docs](https://supabase.com/docs/guides/storage)
- [JWT Token Testing](https://jwt.io/)

---

**🎯 Quick Start:**
1. `docker-compose up -d`
2. `make migrate-up`
3. Buka https://hoppscotch.io
4. Import `scripts/hoppscotch-web-collections.json`
5. Test register → login → upload foto → get profile
