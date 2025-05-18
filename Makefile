# Cloud Drive Backend Makefile
.PHONY: proto build up down clean all

# Tạo mã từ proto definitions
proto:
	@echo "=== Generating code from proto files ==="
	@find proto-definitions -name "*.proto" -exec protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative {} \;
	@echo "=== Proto generation completed ==="

# Build Docker images
build: proto
	@echo "=== Building Docker images ==="
	@cd deployments/docker && docker-compose build
	@echo "=== Docker build completed ==="

# Chạy Docker containers
up:
	@echo "=== Starting Docker containers ==="
	@cd deployments/docker && docker-compose up -d
	@echo "=== Docker containers started ==="

# Dừng Docker containers
down:
	@echo "=== Stopping Docker containers ==="
	@cd deployments/docker && docker-compose down
	@echo "=== Docker containers stopped ==="

# Hiển thị logs
logs:
	@cd deployments/docker && docker-compose logs -f

# Dọn dẹp
clean:
	@echo "=== Cleaning project ==="
	@cd deployments/docker && docker-compose down -v
	@docker system prune -f
	@echo "=== Cleanup completed ==="

# Chạy tất cả
all: proto build up

# Rebuid và khởi động lại
rebuild: down build up

# Help
help:
	@echo "Available commands:"
	@echo "  make proto     - Generate code from proto files"
	@echo "  make build     - Build Docker images"
	@echo "  make up        - Start Docker containers"
	@echo "  make down      - Stop Docker containers"
	@echo "  make logs      - Show Docker logs"
	@echo "  make clean     - Clean up Docker resources"
	@echo "  make all       - Generate proto, build, and start containers"
	@echo "  make rebuild   - Rebuild and restart containers" 