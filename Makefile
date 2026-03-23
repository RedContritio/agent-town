SHELL := /bin/bash -lc

.PHONY: build server cli web debug-server run stop clean test help

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

# Build debug server
debug-server:
	@mkdir -p bin
	@echo "→ Building debug server..."
	@cd .agents/skills/godot-web-debug/server && /usr/local/go/bin/go build -o ../../../../bin/debug-server .
	@echo "✓ bin/debug-server"

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
	@-pkill -f "bin/debug-server" 2>/dev/null || true
	@echo "✓ Server stopped"

# Run debug server only
debug-server-run: debug-server
	@echo "→ Starting debug server on :8081..."
	@./bin/debug-server > /tmp/debug-server.log 2>&1 &
	@sleep 1
	@echo "✓ Debug server running at http://localhost:8081"

# Start Godot directly (no web export needed)
godot-run:
	@echo "→ Starting Godot..."
	@cd godot-web && godot4 . &
	@echo "✓ Godot started"

# Full debug workflow (Godot native mode, no web export)
debug: debug-server server
	@.agents/skills/godot-web-debug/bin/debug-start.sh

# Stop debug environment
debug-stop:
	@.agents/skills/godot-web-debug/bin/debug-stop.sh

# =============================================================================
# Utils
# =============================================================================

clean:
	@rm -rf bin/ server/cmd/server/web/
	@rm -f .agents/skills/godot-web-debug/server/go.sum
	@echo "✓ Cleaned"

test:
	@go test -v ./...

help:
	@echo "Agent Town - Build Commands"
	@echo ""
	@echo "Build:"
	@echo "  make build         # Build everything (server + cli + web)"
	@echo "  make server        # Build server binary only"
	@echo "  make cli           # Build CLI binary only"
	@echo "  make web           # Export web frontend only (requires godot4)"
	@echo "  make debug-server  # Build debug server for godot-web"
	@echo ""
	@echo "Run:"
	@echo "  make run           # Build and start server"
	@echo "  make stop          # Stop server"
	@echo "  make godot-run     # Start Godot only (manual debugging)"
	@echo "  make debug         # Full debug environment (servers + Godot)"
	@echo "  make debug-stop    # Stop debug environment"
	@echo ""
	@echo "Utils:"
	@echo "  make clean         # Clean build artifacts"
	@echo "  make test          # Run tests"
	@echo "  make help          # Show this help"
	@echo ""
	@echo "Setup:    ./scripts/setup.sh"
	@echo "Monitor:  ./scripts/dev.sh {status|logs|test|env}"
