.PHONY: build server cli web run stop clean test help

# =============================================================================
# Build
# =============================================================================

# Build all (server + cli + web)
build: server cli web

# Build server binary
server:
	@mkdir -p bin
	@echo "→ Building server..."
	@go build -o bin/server ./server/cmd/server
	@echo "✓ bin/server"

# Build CLI binary
cli:
	@mkdir -p bin
	@echo "→ Building cli..."
	@go build -o bin/cli ./cli/cmd/cli
	@echo "✓ bin/cli"

# Export web frontend (Godot WebAssembly)
web:
	@echo "→ Exporting web frontend..."
	@mkdir -p server/cmd/server/web
	@cd godot-web && godot4 --headless --export-release "Web" ../server/cmd/server/web/index.html
	@echo "✓ Web exported"

# =============================================================================
# Run
# =============================================================================

run: server
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
	@echo "Build:"
	@echo "  make build    # Build everything (server + cli + web)"
	@echo "  make server   # Build server binary only"
	@echo "  make cli      # Build CLI binary only"
	@echo "  make web      # Export web frontend only (requires godot4)"
	@echo ""
	@echo "Run:"
	@echo "  make run      # Build and start server"
	@echo "  make stop     # Stop server"
	@echo ""
	@echo "Utils:"
	@echo "  make clean    # Clean build artifacts"
	@echo "  make test     # Run tests"
	@echo "  make help     # Show this help"
	@echo ""
	@echo "Setup:    ./scripts/setup.sh"
	@echo "Monitor:  ./scripts/dev.sh {status|logs|test|env}"
