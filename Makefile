.PHONY: all build proto clean test server cli web web-deps

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

# Node settings
NPM=npm

all: proto build

# Build all binaries
build: server cli web

# Build server binary
server:
	cd server/cmd/server && $(GOBUILD) -o ../../../bin/server .

# Build CLI binary
cli:
	cd cli/cmd/cli && $(GOBUILD) -o ../../../bin/cli .

# Build web (production)
web:
	cd web && $(NPM) run build

# Install web dependencies
web-deps:
	cd web && $(NPM) install

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
	rm -rf web/dist/
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

dev-web:
	cd web && $(NPM) run dev

# Docker commands
docker-build:
	docker-compose build

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
	@echo "  make web          - Build web (production)"
	@echo "  make web-deps     - Install web dependencies"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make deps         - Download Go dependencies"
	@echo "  make dev-server   - Run server in development mode"
	@echo "  make dev-cli      - Run CLI in development mode"
	@echo "  make dev-web      - Run web in development mode"
