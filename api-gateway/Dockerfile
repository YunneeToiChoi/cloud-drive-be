FROM golang:1.23 AS builder

WORKDIR /build

# Copy proto-definitions trước
COPY proto-definitions/ /build/proto-definitions/

# Copy api-gateway
COPY api-gateway/ /build/api-gateway/

WORKDIR /build/api-gateway

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o app ./cmd/server

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata && \
    update-ca-certificates

WORKDIR /app

COPY --from=builder /build/api-gateway/app /app/

EXPOSE 8000

CMD ["/app/app"] 