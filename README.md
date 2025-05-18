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

## Chạy hệ thống

### 1. Chạy với Docker (Khuyến nghị cho phát triển)

```bash
cd deployments/docker
docker-compose up -d
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

## API Endpoints

### API Gateway (Port mặc định: 8080 cho local, 8000 cho Docker)

#### Health Check
```
GET /health
```

#### User Service APIs
```
GET    /api/users            # Lấy danh sách người dùng
POST   /api/users            # Tạo người dùng mới
GET    /api/users/{id}       # Lấy thông tin người dùng theo ID
PUT    /api/users/{id}       # Cập nhật thông tin người dùng
DELETE /api/users/{id}       # Xóa người dùng
```

### User Service gRPC API (Port mặc định: 9001)

```
service UserService {
  rpc CreateUser(CreateUserRequest) returns (UserResponse);
  rpc GetUser(GetUserRequest) returns (UserResponse);
  rpc UpdateUser(UpdateUserRequest) returns (UserResponse);
  rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse);
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
}
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
