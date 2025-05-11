# Dockerfile
FROM golang:1.22-alpine AS builder

# Tạo thư mục làm việc
WORKDIR /app

# Copy toàn bộ mã nguồn vào container
COPY . .

# Tải dependencies và build
WORKDIR /app/api-gateway
RUN go mod download
RUN go build -o gateway main.go

# Stage 2: Build final image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/api-gateway/gateway .
EXPOSE 8080
CMD ["./gateway"]
