FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy workspace file và các module
COPY go.work go.work
COPY user user
COPY common-lib common-lib
COPY api-gateway api-gateway

# Đồng bộ workspace để các module nhìn thấy nhau
RUN go work sync

# Chạy tidy cho từng module để tạo đầy đủ go.sum
WORKDIR /app/common-lib
RUN go mod tidy

WORKDIR /app/user
RUN go mod tidy
RUN go build -o user-service main.go

# Stage 2: Build final image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/user/user-service .
EXPOSE 8081
CMD ["./user-service"]
