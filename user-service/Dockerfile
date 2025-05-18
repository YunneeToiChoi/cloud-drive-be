FROM golang:1.23 AS builder

# Tạo thư mục để chứa toàn bộ mã nguồn
WORKDIR /build

# Copy mã nguồn proto-definitions trước
COPY proto-definitions/ /build/proto-definitions/

# Copy mã nguồn user-service
COPY user-service/ /build/user-service/

# Di chuyển vào thư mục service để build
WORKDIR /build/user-service

# Build với cờ tối ưu hóa cho production
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o app ./cmd/server

# Image nhẹ hơn cho production
FROM alpine:3.19

# Cài đặt ca-certificates cho TLS
RUN apk --no-cache add ca-certificates tzdata && \
    update-ca-certificates

# Tạo thư mục không đặc quyền
WORKDIR /app

# Copy chỉ file thực thi
COPY --from=builder /build/user-service/app /app/

# Mở port cần thiết
EXPOSE 9001

# Chạy ứng dụng
CMD ["/app/app"] 