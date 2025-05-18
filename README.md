# Cloud Drive Backend

Hệ thống microservices cho Cloud Drive sử dụng Golang, gRPC, và Consul.

## Cấu trúc dự án

```
cloud-drive-be/
├── api-gateway/         # API Gateway (Go)
├── user-service/        # User Service (Go)
├── proto-definitions/   # Shared Proto Definitions
├── shared/              # Shared Code
├── deployments/         # Deployment Configurations
│   ├── docker/          # Docker Compose
│   └── nginx/           # Nginx Configuration
└── docs/                # Documentation
```

## Các port mặc định

| Service         | Local              | Docker              | Production         |
|-----------------|--------------------|--------------------|-------------------|
| API Gateway     | 8080               | 8000               | 80                |
| User Service    | 9001 (gRPC)        | 9001 (gRPC)        | 9001 (gRPC)       |
| Consul          | 8500 (HTTP), 8600 (DNS) | 8500 (HTTP), 8600 (DNS) | 8500 (HTTP), 8600 (DNS) |
| PostgreSQL      | 5432               | 5432               | 5432              |

## Cài đặt

### Yêu cầu

- Go 1.22+
- Docker & Docker Compose
- Git
- [Protocol Buffers Compiler](https://grpc.io/docs/protoc-installation/)

### Clone repository

```bash
git clone https://github.com/organization/cloud-drive-be.git
cd cloud-drive-be
git submodule update --init --recursive
```

### Cài đặt dependencies

```bash
go mod download
```

## Hướng dẫn sử dụng Protocol Buffers

### Cài đặt Protobuf Compiler

#### Windows
```bash
# Sử dụng Chocolatey
choco install protoc

# Cài đặt Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

#### Linux/WSL
```bash
# Cài đặt protoc
apt update
apt install -y protobuf-compiler

# Cài đặt Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

#### macOS
```bash
# Sử dụng Homebrew
brew install protobuf

# Cài đặt Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Cập nhật proto definitions

1. Chỉnh sửa file proto trong thư mục `proto-definitions/`
2. Tạo mã từ proto definitions:

```bash
# Sử dụng Makefile (phải chạy từ thư mục gốc của dự án)
cd /path/to/cloud-drive-be
make proto

# Hoặc chạy lệnh protoc trực tiếp (khi Makefile không hoạt động)
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto-definitions/user/user.proto
```

**Lưu ý**: Khi sử dụng `make proto`, luôn đảm bảo bạn đang ở thư mục gốc của dự án.

## Hướng dẫn sử dụng Makefile

Dự án sử dụng Makefile để tự động hóa quy trình phát triển. Dưới đây là các lệnh make có sẵn:

```bash
# Tạo mã từ proto definitions
make proto

# Build Docker images
make build

# Chạy tất cả các service
make up

# Dừng tất cả các service
make down

# Xem log
make logs

# Dọn dẹp (xóa volumes và containers)
make clean

# Tất cả các bước: tạo proto, build và chạy
make all

# Rebuild và khởi động lại
make rebuild
```

### Windows

Để sử dụng Makefile trên Windows, bạn có thể:

1. Sử dụng WSL (Windows Subsystem for Linux)
2. Cài đặt Make thông qua Chocolatey: `choco install make`
3. Sử dụng Docker Desktop có tích hợp Make

## Chạy hệ thống

### 1. Chạy với Docker (Khuyến nghị cho phát triển)

```bash
cd deployments/docker
docker-compose up -d
```

**RESET DOCKER**
```bash
cd deployments/docker
docker stop $(docker ps -aq) 2>/dev/null || true && \
docker rm -f $(docker ps -aq) 2>/dev/null || true && \
docker rmi -f $(docker images -aq) 2>/dev/null || true && \
docker builder prune -a -f && \
docker-compose   build --no-cache && \
docker-compose  up -d --force-recreate
```
Kiểm tra trạng thái:
```bash
docker-compose ps
```

### 2. Chạy Local (Debug)

**Bước 1:** Chạy Consul và PostgreSQL bằng Docker:
```bash
cd deployments/docker
docker-compose up -d consul postgres
```

**Bước 2:** Cấu hình biến môi trường (tạo file .env từ .env.example):
```bash
cp .env.example .env
```

**Bước 3:** Chạy API Gateway:
```bash
cd api-gateway
go run cmd/server/main.go
```

**Bước 4:** Chạy User Service trong terminal khác:
```bash
cd user-service
go run cmd/server/main.go
```

### 3. Chạy trong Production

Cấu hình production được quản lý qua các biến môi trường `PROD_*` trong file `.env`.

```bash
# Đặt môi trường thành production
export APP_ENV=production

# Chạy Docker với môi trường production
docker-compose -f deployments/docker/docker-compose.yml up -d
```

## API Endpoints và JWT Authentication

### Xác thực JWT

Hệ thống sử dụng JSON Web Token (JWT) cho xác thực. Middleware JWT được triển khai trong `api-gateway/internal/middleware/auth.go`.

#### Các endpoint không yêu cầu xác thực:
```
POST   /api/auth/login      # Đăng nhập
POST   /api/auth/register   # Đăng ký tài khoản mới
GET    /health              # Kiểm tra trạng thái API Gateway
```

#### Các endpoint yêu cầu xác thực JWT:
```
GET    /api/users               # Lấy danh sách người dùng (yêu cầu quyền admin)
POST   /api/users               # Tạo người dùng mới (yêu cầu quyền admin)
GET    /api/users/{id}          # Lấy thông tin người dùng (người dùng chỉ xem được thông tin của chính mình)
PUT    /api/users/{id}          # Cập nhật thông tin người dùng (người dùng chỉ cập nhật được thông tin của chính mình)
DELETE /api/users/{id}          # Xóa người dùng (yêu cầu quyền admin)
```

### Workflow và cách kiểm tra JWT

#### 1. Đăng ký tài khoản:
```
POST /api/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword",
  "first_name": "John",
  "last_name": "Doe"
}
```

#### 2. Đăng nhập để lấy token:
```
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword"
}
```

Phản hồi:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 86400,
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "role": "user"
}
```

#### 3. Sử dụng token để gọi API được bảo vệ:
```
GET /api/users/123e4567-e89b-12d3-a456-426614174000
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Kiểm tra middleware và JWT

Để kiểm tra xem JWT middleware có hoạt động đúng cách không, bạn có thể sử dụng các endpoint sau:

#### Kiểm tra xác thực:
```bash
# Không có token - sẽ trả về 401 Unauthorized
curl -i http://localhost:8080/api/users/123

# Với token không hợp lệ - sẽ trả về 401 Unauthorized
curl -i -H "Authorization: Bearer invalid-token" http://localhost:8080/api/users/123

# Với token hợp lệ - sẽ trả về dữ liệu hoặc 403 nếu không có quyền
curl -i -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." http://localhost:8080/api/users/123
```

#### Kiểm tra phân quyền:
```bash
# User thông thường truy cập thông tin người khác - sẽ trả về 403 Forbidden
curl -i -H "Authorization: Bearer [token-user]" http://localhost:8080/api/users/different-id

# Admin truy cập thông tin bất kỳ - sẽ thành công
curl -i -H "Authorization: Bearer [token-admin]" http://localhost:8080/api/users/any-id

# User thông thường truy cập /api/users - sẽ trả về 403 Forbidden
curl -i -H "Authorization: Bearer [token-user]" http://localhost:8080/api/users
```

## Xác thực và Phân quyền

### Cơ chế JWT
Hệ thống sử dụng JWT cho xác thực người dùng. Token được tạo khi người dùng đăng nhập thành công và phải được gửi trong header `Authorization` của các request.

### Headers
```
Authorization: Bearer <token>
```

### Roles và Permissions
Hệ thống hỗ trợ các vai trò (roles) khác nhau:
- **user**: Người dùng thông thường, chỉ có thể xem và cập nhật thông tin của chính mình
- **admin**: Quản trị viên, có quyền truy cập tất cả tài nguyên

### Biến môi trường JWT
```
DEV_JWT_SECRET=dev_jwt_secret_key
PROD_JWT_SECRET=prod_jwt_secret_key
JWT_EXPIRATION=24h
```

## Môi trường và lưu ý

### Local Development

- **Biến môi trường**: `APP_ENV=development`, `HOST_MODE=local`
- **URL Consul**: `127.0.0.1:8500` (mặc định)
- **Lưu ý**:
  - Cần chạy Consul và PostgreSQL trước khi khởi động các services
  - Service sẽ tự động phát hiện đang chạy local và kết nối qua `127.0.0.1`
  - Sử dụng GoLand hoặc VSCode để debug (cung cấp biến môi trường trong cấu hình chạy)

### Docker Development

- **Biến môi trường**: `APP_ENV=development`, `HOST_MODE=docker`
- **URL Consul**: `consul:8500` (tên service trong mạng Docker)
- **Lưu ý**:
  - Tất cả service đều kết nối với nhau qua tên service
  - Sử dụng `docker-compose logs -f <service>` để xem logs
  - Có thể cập nhật và rebuild service cụ thể: `docker-compose build user-service`

### Production

- **Biến môi trường**: `APP_ENV=production`, `HOST_MODE=docker`
- **URL Consul**: `consul:8500` (tên service trong mạng Docker)
- **Lưu ý**:
  - Cấu hình chi tiết thông qua biến môi trường `PROD_*` 
  - Đảm bảo sử dụng mật khẩu mạnh trong production
  - Xem xét sử dụng Kubernetes thay vì Docker Compose
  - Cấu hình SSL/TLS với Nginx trong production
  - Sử dụng mật khẩu mạnh cho `PROD_JWT_SECRET`

## Debug trong GoLand

1. Chạy Consul và PostgreSQL với Docker
2. Tạo cấu hình chạy mới trong GoLand
3. Thêm biến môi trường:
   ```
   APP_ENV=development;HOST_MODE=local
   ```
4. Run hoặc Debug service từ IDE

## Tự động phát hiện môi trường

Hệ thống tự động phát hiện nếu đang chạy trong Docker hay local qua:
- Biến môi trường `HOST_MODE`
- Hoặc kiểm tra file `/.dockerenv` (chỉ tồn tại trong Docker)

## Cấu hình Service Discovery

Các service tự động tìm kiếm nhau qua Consul, không cần cấu hình URL cứng. URL Consul được tự động phát hiện dựa trên môi trường:
- Docker: `consul:8500`
- Local: `127.0.0.1:8500`
