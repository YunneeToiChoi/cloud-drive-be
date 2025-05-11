.PHONY: build run docker docker-build docker-run docker-stop clean

# Build tất cả các services
build:
	cd api-gateway && go build -o bin/gateway main.go
	cd user && go build -o bin/user-service main.go

# Chạy các services (không dùng Docker)
run-api-gateway:
	cd api-gateway && go run main.go

run-user:
	cd user && go run main.go

# Docker commands
docker-build:
	docker-compose build

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down

# Làm sạch
clean:
	rm -rf api-gateway/bin
	rm -rf user/bin
	
# Chạy tất cả trong Docker
all: docker-build docker-run
	@echo "All services are running"

# Hiển thị logs
logs:
	docker-compose logs -f
