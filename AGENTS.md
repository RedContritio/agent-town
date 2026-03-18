# Agent-Town Project Guide

This document provides essential information for AI coding agents working on the Agent-Town project.

## Project Overview

Agent-Town is a 3D virtual town sandbox simulation game where:
- **AI Agents** interact with the world via CLI (automated or human-controlled)
- **Human Observers** watch and lightly intervene via Web browser
- **Server** manages the world state, physics, economy, and agent interactions

The project aims to create a self-evolving society where agents autonomously build, trade, and form social structures.

## Technology Stack

| Component | Technology | Version |
|-----------|------------|---------|
| Backend | Go | 1.22+ |
| Web Frontend | Godot | 4.5+ |
| 3D Rendering | Godot Engine | Built-in |
| Web Export | WebAssembly | WASM |
| API Protocol | HTTP REST | JSON |
| Database | PostgreSQL | 16 (planned) |
| Cache | Redis | 7 (planned) |
| Containerization | Docker Compose | 3.8 |

## Project Structure

```
agent-town/
├── cli/                    # CLI client for agents (Go)
│   └── cmd/cli/
│       └── main.go         # Entry point (placeholder)
├── server/                 # Backend server (Go)
│   └── cmd/server/
│       ├── main.go         # HTTP server with REST API
│       └── web/            # Godot web export (generated, gitignored)
├── godot-web/              # Web frontend (Godot project)
│   ├── scenes/             # Godot scene files
│   │   ├── entities/       # Agent, Building scenes
│   │   ├── ui/             # HUD, UI scenes
│   │   └── main.tscn       # Main scene
│   ├── scripts/            # GDScript files
│   │   ├── api_client.gd   # HTTP API client
│   │   ├── camera_controller.gd
│   │   ├── world_manager.gd
│   │   └── entities/       # Entity scripts
│   ├── project.godot       # Godot project config
│   └── export_presets.cfg  # Export configuration
├── proto/                  # Protocol Buffer definitions (planned for gRPC)
│   └── agenttown/v1/
│       ├── agent.proto
│       └── common.proto
├── docs/                   # Documentation
│   ├── GAME_DESIGN.md      # Game design document (Chinese)
│   ├── TECH_DESIGN.md      # Technical design document (Chinese)
│   ├── WEB_DESIGN.md       # Web UI/UX design (legacy, see VISUAL_DESIGN.md)
│   └── VISUAL_DESIGN.md    # Visual specification
├── docker-compose.yml      # Docker services configuration
├── Makefile               # Build automation
├── go.mod                 # Go module definition
└── AGENTS.md              # This file
```

## Build Commands

The project uses Make for build automation:

```bash
# Build everything (proto + binaries)
make all

# Build specific components
make server        # Build server binary to bin/server
make cli           # Build CLI binary to bin/cli
make godot-web     # Export Godot web frontend to server/cmd/server/web/

# Development mode
make dev-server    # Run server with go run
make dev-cli       # Run CLI with go run

# Code generation
make proto         # Generate Go code from .proto files

# Testing
make test          # Run Go tests

# Dependencies
make deps          # Download Go dependencies

# Cleanup
make clean         # Remove build artifacts

# Docker
make docker-up     # Start services with docker-compose
make docker-down   # Stop services
```

## Development Workflow

### 1. Initial Setup

```bash
# 1. Install Go 1.22+
# 2. Install Godot 4.5+ (for web export)
# 3. Install Protocol Buffer compiler (protoc) - optional

# 4. Download dependencies
make deps

# 5. Start infrastructure services (optional, currently using mock data)
make docker-up

# 6. Export Godot web frontend and run server
make godot-web
make dev-server

# 7. Open browser at http://localhost:8080
```

### 2. Godot Web Development

When modifying Godot scenes or scripts:

```bash
# Export web version
cd godot-web && godot --headless --export-release "Web" ../server/cmd/server/web/index.html

# Or use make
make godot-web
```

Godot project structure:
- `scenes/` - Scene files (.tscn)
- `scripts/` - GDScript files (.gd)
- `assets/` - 3D models, textures, materials

### 3. Protocol Buffer Changes (Future)

When modifying `.proto` files:

```bash
make proto         # Regenerate Go code
```

Generated files are placed in `proto/agenttown/v1/*.pb.go`.

## Architecture Details

### Current Implementation State

**Server** (`server/cmd/server/main.go`):
- Currently a mock HTTP server serving hardcoded JSON data
- Implements REST API endpoints under `/api/v1/`
- Serves static web files from `./web` (Godot export)
- CORS enabled for all origins
- Returns proper MIME types and COOP/COEP headers for Godot Web

**CLI** (`cli/cmd/cli/main.go`):
- Placeholder implementation
- Intended for agent registration, login, and command execution

**Godot Web** (`godot-web/`):
- Fully functional 3D world viewer
- Uses Godot Engine 4.5+ with WebAssembly export
- Renders terrain (MultiMesh for performance), buildings, and agents
- Auto-refreshes world data every 5 seconds via HTTP API
- HUD as separate CanvasLayer for UI elements
- Name labels in screen space (not affected by 3D occlusion)

### API Endpoints (Current)

```
GET  /api/v1/world/info       # World metadata
GET  /api/v1/world/time       # World time
GET  /api/v1/world/map        # Map data (query: x, y, radius)
GET  /api/v1/agents           # List all agents
GET  /api/v1/agents/{id}      # Agent details
GET  /api/v1/agents/{id}/status
GET  /api/v1/agents/{id}/visible-area
GET  /api/v1/agents/{id}/todos
GET  /api/v1/agents/{id}/skills
```

### Planned gRPC Services (from proto/)

- `AgentService` - Agent management, authentication, status
- `ActionService` - Gather, build, farm, talk, trade, battle
- `BattleService` - PvP/PvE combat
- `TaskService` - Quest system
- `WorldService` - World info, chunks, time
- `EconomyService` - Transfers, transactions, friends

## Code Style Guidelines

### Go
- Follow standard Go formatting (`gofmt`)
- Use `go vet` for static analysis
- Prefer explicit error handling
- Package structure: `cmd/` for executables, `internal/` for private code

### GDScript (Godot)
- Use `snake_case` for functions and variables
- Use `PascalCase` for class names and constants
- Indent with tabs (Godot convention)
- Type hints encouraged for function parameters and returns
- Signal names use `snake_case` with past tense verbs (`player_died`)

### Protocol Buffers
- Use proto3 syntax
- Package: `agenttown.v1`
- Go package: `github.com/RedContritio/agent-town/proto/agenttown/v1`
- Messages use snake_case field names

## Testing Strategy

```bash
# Run all Go tests
make test

# Run specific package
go test -v ./server/...
```

**Note**: The project currently has minimal test coverage. Tests should be added for:
- Core game logic (world generation, agent actions)
- API handlers
- Protocol buffer message validation

## Key Configuration

### Environment Variables (Server)

```bash
DATABASE_URL=postgres://agenttown:agenttown@localhost:5432/agenttown?sslmode=disable
REDIS_URL=localhost:6379
GRPC_PORT=50051
HTTP_PORT=8080
```

### Godot Export

Export settings in `godot-web/export_presets.cfg`:
- Export path: `../server/cmd/server/web/index.html`
- Custom HTML shell: Default
- Head include: None (headers set by Go server)

## Docker Services

The `docker-compose.yml` defines:

| Service | Image | Port | Purpose |
|---------|-------|------|---------|
| postgres | postgres:16-alpine | 5432 | Main database |
| redis | redis:7-alpine | 6379 | Cache & sessions |

**Note**: Currently using mock data; database not actively used.

## GitIgnore Structure

Each directory has its own `.gitignore`:

- `.gitignore` (root) - IDE, OS, environment variables, logs
- `godot-web/.gitignore` - `.godot/`, `.import/`, Godot cache
- `server/cmd/server/.gitignore` - Web export artifacts (`web/`, `*.wasm`, `*.pck`)

## Visual Design

See `docs/VISUAL_DESIGN.md` for complete visual specification including:
- Color palette (UI and world)
- Typography
- Building type colors
- UI component styles

Key colors:
- UI Background: `#1a1a2e`
- UI Accent: `#4fc3f7`
- Sky: `#87ceeb`
- Ground: `#c0c0c0`

## Important Design Decisions

1. **TODO Token System**: Agents generate tokens via CLI to authorize web users to manage their TODO list. Token passed via `X-Todo-Token` header.

2. **Vision System**: Server only provides data within agent's field of view. Agents must maintain their own "mental map" of explored areas.

3. **Identity**: Agents use public-key cryptography for authentication. Agent ID derived from public key.

4. **Economy**: All transactions visible to all agents (transparent economy). Taxes adjustable by referendum.

5. **Combat**: Turn-based system. PvP requires mutual consent. PvE allows offline stalemate.

## Documentation References

All detailed design documents are in `docs/` directory:

- `GAME_DESIGN.md` - Game mechanics, world systems, agent lifecycle (Chinese)
- `TECH_DESIGN.md` - Database schema, API design, service architecture (Chinese)
- `WEB_DESIGN.md` - Legacy web design (React era), see VISUAL_DESIGN.md for current
- `VISUAL_DESIGN.md` - Current visual specification

## Common Tasks

### Adding a New API Endpoint

1. Add HTTP handler in `server/cmd/server/main.go`
2. Update `godot-web/scripts/api_client.gd` to add client method
3. Use the new API in appropriate scene/script

### Adding a New Terrain Type

1. Update terrain generation logic in server (when implemented)
2. Add material definition in `godot-web/scripts/world_manager.gd`
3. Update color in `docs/VISUAL_DESIGN.md`

### Adding a New Building Type

1. Update `buildingColors` in `godot-web/scripts/world_manager.gd`
2. Add to legend UI in `godot-web/scenes/ui/hud.tscn`
3. Update color specification in `docs/VISUAL_DESIGN.md`

### Database Schema Changes

1. Update schema in `docs/TECH_DESIGN.md`
2. Create migration files (when migration system is implemented)
3. Update Go structs
4. Update proto messages if API changes

## Security Considerations

- **Authentication**: JWT tokens after public-key signature verification
- **TODO Token**: Separate scoped token for web TODO management
- **Rate Limiting**: Should be implemented on API endpoints (currently TODO)
- **Input Validation**: Validate all incoming JSON data
- **CORS**: Currently allows all origins (`*`) - restrict in production
- **COOP/COEP**: Required headers set for Godot WebAssembly (SharedArrayBuffer)

## License

MIT License - Copyright (c) 2026 RedContritio
