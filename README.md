# Microservices Demo Project

Đây là demo cho kiến trúc microservices sử dụng Go, Docker và Consul Service Discovery.

## Cấu trúc Project

Project bao gồm các thành phần chính:

- **API Gateway**: Cổng vào cho tất cả các request từ client
- **Users Service**: Quản lý thông tin người dùng
- **Common Lib**: Thư viện chung cho các service
- **Consul**: Dịch vụ service discovery

## Cài đặt và Chạy

### Yêu cầu

- Docker và Docker Compose
- Go 1.20+

### Các bước chạy project

1. **Clone repository**

2. **Build và chạy với Docker**

```bash
docker-compose build
docker-compose up
```

3. **Truy cập các dịch vụ**

- API Gateway: http://localhost:8080
- Users Service: http://localhost:8081
- Consul UI: http://localhost:8500

## API Endpoints

### 1. Lấy thông tin người dùng

```
GET http://localhost:8080/api/users/GetById?userId=1
```

**Tham số:**
- `userId`: ID của người dùng cần lấy thông tin

**Response mẫu:**
```json
{
  "id": "1",
  "name": "John Doe",
  "email": "john.doe@example.com"
}
```

### 2. Kiểm tra health của Users Service

```
GET http://localhost:8080/api/users/health
```

**Response mẫu:**
```json
{
  "status": "User service is healthy"
}
```

### 3. Kiểm tra trực tiếp Users Service (không qua API Gateway)

```
GET http://localhost:8081/api/users/health
```

## Cấu hình Postman

1. **Tạo Collection mới** cho project

2. **Thêm Request mới**:
   - **Name**: Get User By ID
   - **Method**: GET
   - **URL**: http://localhost:8080/api/users/GetById?userId=1

3. **Thêm Health Check Request**:
   - **Name**: Users Service Health
   - **Method**: GET
   - **URL**: http://localhost:8080/api/users/health

## Kiến trúc

### Service Discovery

Project sử dụng Consul làm service registry. Các service đăng ký với Consul khi khởi động và API Gateway tìm kiếm service thông qua Consul. Nếu không kết nối được với Consul, hệ thống sẽ fallback sang sử dụng biến môi trường.

### Clean Architecture

Project tuân theo nguyên tắc Clean Architecture với các layer rõ ràng:
- **Handler**: Xử lý HTTP request/response
- **Service**: Chứa business logic
- **Repository**: Tương tác với dữ liệu

## Phát triển Local

Để chạy project trên môi trường phát triển local mà không dùng Docker:

1. Chạy Consul với Docker hoặc cài đặt trực tiếp

```bash
docker run -d -p 8500:8500 consul:1.14 agent -server -ui -node=server-1 -bootstrap-expect=1 -client=0.0.0.0
```

2. chạy Goland ( config build all hoặc build từng services lên)

nếu không chạy được bằng Goland ( thực hiện các bước dưới )

2. Chạy Users Service

```bash
cd user
go run main.go
```

3. Chạy API Gateway

```bash
cd api-gateway
go run main.go
```

## Troubleshooting

### Lỗi Port đã được sử dụng

Nếu gặp lỗi port đã được sử dụng, bạn có thể:

1. Tắt tiến trình đang sử dụng port:
```
netstat -ano | findstr <PORT>
taskkill /PID <PID> /F
```

2. Hoặc thay đổi port mapping trong docker-compose.yml
