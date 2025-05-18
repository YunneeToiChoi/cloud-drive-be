# Cloud Drive Backend

Dự án backend sử dụng microservices architecture với Golang.

## Cấu trúc dự án
```
├── api-gateway/          # API Gateway service
├── user-service/         # User service
├── discovery-service/    # Service discovery
├── shared/               # Shared code và utilities
├── deployments/          # Kubernetes, Docker configs
└── docs/                 # Documentation
```

## Các thành phần
- API Gateway: Điểm vào cho các requests
- User Service: Quản lý thông tin người dùng
- Discovery Service: Service discovery sử dụng Consul

## Cách chạy
Chi tiết setup và running các services có trong thư mục tương ứng
