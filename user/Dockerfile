FROM golang:1.22-alpine AS builder

# Tạo thư mục làm việc
WORKDIR /app

# Copy toàn bộ mã nguồn vào container
COPY . .

# Tải dependencies và build
WORKDIR /app/user
RUN go mod download
RUN go build -o user-service main.go

# Stage 2: Build final image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/user/user-service .
# Nếu có file env thì copy ở đây
EXPOSE 8081
CMD ["./user-service"]
