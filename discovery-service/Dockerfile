FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o discovery-service ./cmd/server

# Use a minimal alpine image
FROM alpine:latest

RUN apk --no-cache add ca-certificates && update-ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/discovery-service .

# Expose the port
EXPOSE 8500

# Run the application
CMD ["./discovery-service"] 