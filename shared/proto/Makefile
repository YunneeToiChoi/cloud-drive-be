.PHONY: gen-proto clean

# Generate Go code from proto files
gen-proto:
	@echo "Generating Go code from proto files..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		user/user.proto
	@echo "Done."

# Copy generated files to services
copy-to-services: gen-proto
	@echo "Copying generated files to services..."
	mkdir -p ../../api-gateway/proto/user
	cp user/*.pb.go ../../api-gateway/proto/user/
	mkdir -p ../../user-service/proto/user
	cp user/*.pb.go ../../user-service/proto/user/
	@echo "Done."

# Clean generated files
clean:
	@echo "Cleaning generated files..."
	rm -f user/*.pb.go
	@echo "Done." 