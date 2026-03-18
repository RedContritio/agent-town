.PHONY: all build proto clean test server cli godot-web docker-up docker-down

# Go settings
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Proto settings
PROTOC=protoc
PROTO_DIR=proto
PROTO_OUT=proto

# Godot settings
GODOT=godot
GODOT_PROJECT=godot-web

all: proto build

# Build all binaries
build: server cli

# Build server binary
server:
	cd server/cmd/server && $(GOBUILD) -o ../../../bin/server .

# Build CLI binary
cli:
	cd cli/cmd/cli && $(GOBUILD) -o ../../../bin/cli .

# Export Godot web frontend
godot-web:
	mkdir -p server/cmd/server/web
	cd $(GODOT_PROJECT) && $(GODOT) --headless --export-release "Web" ../server/cmd/server/web/index.html

# Install web dependencies (legacy, kept for compatibility)
web-deps:
	@echo "Web frontend now uses Godot. No npm dependencies needed."
	@echo "Use 'make godot-web' to export the web frontend."

# Generate protobuf code
proto:
	@echo "Generating protobuf code..."
	@mkdir -p $(PROTO_OUT)/agenttown/v1
	$(PROTOC) --go_out=$(PROTO_OUT) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_OUT) --go-grpc_opt=paths=source_relative \
		-I$(PROTO_DIR) \
		$(PROTO_DIR)/agenttown/v1/*.proto

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf bin/
	rm -rf server/cmd/server/web/
	rm -f $(PROTO_OUT)/agenttown/v1/*.pb.go

# Run tests
test:
	$(GOTEST) -v ./...

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Database migrations
migrate-up:
	@echo "Running database migrations up..."
	# TODO: implement migration

migrate-down:
	@echo "Running database migrations down..."
	# TODO: implement migration

# Development commands
dev-server:
	cd server/cmd/server && $(GOCMD) run .

dev-cli:
	cd cli/cmd/cli && $(GOCMD) run .

# Docker commands
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

# Help
help:
	@echo "Available targets:"
	@echo "  make all          - Generate proto and build all binaries"
	@echo "  make proto        - Generate protobuf code"
	@echo "  make build        - Build all binaries"
	@echo "  make server       - Build server binary"
	@echo "  make cli          - Build CLI binary"
	@echo "  make godot-web    - Export Godot web frontend"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make deps         - Download Go dependencies"
	@echo "  make dev-server   - Run server in development mode"
	@echo "  make dev-cli      - Run CLI in development mode"
	@echo "  make docker-up    - Start Docker services"
	@echo "  make docker-down  - Stop Docker services"
