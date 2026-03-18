.PHONY: all build proto clean test server cli godot-web docker-up docker-down

# Go settings - auto-detect Go path
GOCMD := $(shell which go 2>/dev/null || echo /usr/local/go/bin/go)
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Proto settings
PROTOC=protoc
PROTO_DIR=proto
PROTO_OUT=proto

# Godot settings - auto-detect Godot path
GODOT := $(shell which godot4 2>/dev/null || which godot 2>/dev/null || echo /snap/bin/godot4)
GODOT_PROJECT=godot-web
GODOT_WEB_EXPORT_PATH=server/cmd/server/web

all: proto build

# Build all binaries
build: server cli

# Build server binary
server:
	@mkdir -p bin
	cd server/cmd/server && $(GOBUILD) -o ../../../bin/server .

# Build CLI binary
cli:
	@mkdir -p bin
	cd cli/cmd/cli && $(GOBUILD) -o ../../../bin/cli .

# Export Godot web frontend
godot-web: godot-web-check
	@echo "Using Godot: $(GODOT)"
	mkdir -p $(GODOT_WEB_EXPORT_PATH)
	cd $(GODOT_PROJECT) && $(GODOT) --headless --export-release "Web" ../$(GODOT_WEB_EXPORT_PATH)/index.html

# Check Godot export templates
godot-web-check:
	@if [ ! -f "$(HOME)/.local/share/godot/export_templates/4.5.stable/web_release.zip" ] && \
	    [ ! -f "$(HOME)/snap/godot4/current/.local/share/godot/export_templates/4.5.stable/web_release.zip" ]; then \
		echo "ERROR: Godot Web export templates not found!"; \
		echo "Please install templates: $(GODOT) --headless --install-export-templates 4.5.stable"; \
		exit 1; \
	fi
	@echo "Godot export templates found"

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
	cd server/cmd/server && $(GOCMD) run . 2>&1 | tee /tmp/server.log

dev-cli:
	cd cli/cmd/cli && $(GOCMD) run .

# Docker commands
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

# Check tools
env-check:
	@echo "=== Environment Check ==="
	@echo "Go: $(GOCMD)"
	@$(GOCMD) version 2>/dev/null || echo "Go not found!"
	@echo ""
	@echo "Godot: $(GODOT)"
	@$(GODOT) --version 2>/dev/null || echo "Godot not found!"
	@echo ""
	@echo "Protocol Buffers: $(PROTOC)"
	@$(PROTOC) --version 2>/dev/null || echo "protoc not found!"

# Restart server (stop existing, rebuild, start)
restart-server:
	@echo "Restarting server..."
	-pkill -f "bin/server" 2>/dev/null || true
	@sleep 1
	$(MAKE) server
	./bin/server 2>&1 &
	@echo "Server started. Check http://localhost:8080"

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
	@echo "  make restart-server - Restart server (rebuild + run)"
	@echo "  make env-check    - Check tool versions"
	@echo "  make docker-up    - Start Docker services"
	@echo "  make docker-down  - Stop Docker services"
