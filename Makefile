.PHONY: build web run stop clean test help

# =============================================================================
# Build
# =============================================================================

build:
	@mkdir -p bin
	@echo "→ Building server..."
	@go build -o bin/server ./server/cmd/server
	@echo "✓ bin/server"
	@echo "→ Building cli..."
	@go build -o bin/cli ./cli/cmd/cli
	@echo "✓ bin/cli"

web: build
	@echo "→ Exporting web frontend..."
	@mkdir -p server/cmd/server/web
	@cd godot-web && godot4 --headless --export-release "Web" ../server/cmd/server/web/index.html
	@echo "✓ Web exported"

# =============================================================================
# Run
# =============================================================================

run: build
	@echo "→ Starting server..."
	@ln -sf $(PWD)/server/cmd/server/web bin/web 2>/dev/null || true
	@./bin/server > /tmp/agent-town-server.log 2>&1 &
	@sleep 1
	@curl -s http://localhost:8080/api/v1/world/info > /dev/null && \
		echo "✓ Server running at http://localhost:8080" || \
		{ echo "✗ Failed to start"; exit 1; }

stop:
	@-pkill -f "bin/server" 2>/dev/null || true
	@echo "✓ Server stopped"

# =============================================================================
# Utils
# =============================================================================

clean:
	@rm -rf bin/ server/cmd/server/web/
	@echo "✓ Cleaned"

test:
	@go test -v ./...

help:
	@echo "Agent Town - Build Commands"
	@echo ""
	@echo "  make build    # Build server + cli"
	@echo "  make web      # Export web frontend (requires godot4)"
	@echo "  make run      # Start server"
	@echo "  make stop     # Stop server"
	@echo "  make clean    # Clean build artifacts"
	@echo "  make test     # Run tests"
	@echo ""
	@echo "Setup:    ./scripts/setup.sh"
	@echo "Monitor:  ./scripts/dev.sh {status|logs|test|env}"
