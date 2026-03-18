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
| Web Frontend | React + TypeScript | React 18 |
| 3D Rendering | Three.js | 0.160.0 |
| React 3D | @react-three/fiber | 8.15.0 |
| Build Tool | Vite | 5.0.0 |
| State Management | Zustand | 4.4.0 |
| API Protocol | Protocol Buffers + gRPC (planned) | v3 |
| Database | PostgreSQL | 16 |
| Cache | Redis | 7 |
| Containerization | Docker Compose | 3.8 |

## Project Structure

```
agent-town/
├── cli/                    # CLI client for agents (Go)
│   └── cmd/cli/
│       └── main.go         # Entry point (placeholder)
├── server/                 # Backend server (Go)
│   └── cmd/server/
│       └── main.go         # HTTP server with mock data
├── web/                    # Web frontend (React + TypeScript)
│   ├── src/
│   │   ├── api/            # API client (Axios)
│   │   ├── components/     # React components
│   │   │   └── WorldScene.tsx    # Main 3D scene
│   │   ├── stores/         # Zustand state management
│   │   │   ├── worldStore.ts
│   │   │   └── agentStore.ts
│   │   ├── types/          # TypeScript type definitions
│   │   ├── App.tsx
│   │   └── main.tsx
│   ├── package.json
│   ├── tsconfig.json
│   └── vite.config.ts
├── proto/                  # Protocol Buffer definitions
│   └── agenttown/v1/
│       ├── agent.proto     # Agent service definitions
│       └── common.proto    # Common message types
├── docs/                   # Documentation (in Chinese)
│   ├── GAME_DESIGN.md      # Game design document
│   ├── TECH_DESIGN.md      # Technical design document
│   └── WEB_DESIGN.md       # Web UI/UX design document
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
make web           # Build web production bundle

# Development mode
make dev-server    # Run server with go run
make dev-cli       # Run CLI with go run
make dev-web       # Run web dev server (port 3000)

# Code generation
make proto         # Generate Go code from .proto files

# Testing
make test          # Run Go tests

# Dependencies
make deps          # Download Go dependencies
make web-deps      # Install npm dependencies

# Cleanup
make clean         # Remove build artifacts

# Docker
make docker-up     # Start services with docker-compose
make docker-down   # Stop services
```

## Development Workflow

### 1. Initial Setup

```bash
# 1. Install Go 1.22+ and Node.js 18+
# 2. Install Protocol Buffer compiler (protoc)

# 3. Download dependencies
make deps
make web-deps

# 4. Start infrastructure services
make docker-up

# 5. Run in development mode
# Terminal 1: make dev-server
# Terminal 2: make dev-web
```

### 2. Protocol Buffer Changes

When modifying `.proto` files:

```bash
make proto         # Regenerate Go code
```

Generated files are placed in `proto/agenttown/v1/*.pb.go`.

### 3. Web Development

The web frontend uses Vite with hot module replacement:

```bash
cd web
npm run dev        # Development server on port 3000
npm run build      # Production build
npm run preview    # Preview production build
```

API proxy is configured in `vite.config.ts` to forward `/api` to `localhost:8080`.

## Architecture Details

### Current Implementation State

**Server** (`server/cmd/server/main.go`):
- Currently a mock HTTP server serving hardcoded JSON data
- Implements REST API endpoints under `/api/v1/`
- Serves static web files from `./web/dist`
- CORS enabled for all origins

**CLI** (`cli/cmd/cli/main.go`):
- Placeholder implementation
- Intended for agent registration, login, and command execution

**Web** (`web/src/`):
- Fully functional 3D world viewer
- Uses Three.js with React Three Fiber
- Renders terrain (instanced mesh for performance), buildings, and agents
- Auto-refreshes world data every 10 seconds

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

### TypeScript/React
- Strict TypeScript mode enabled
- Use functional components with hooks
- Path alias `@/` maps to `src/`
- Prefer early returns over nested conditionals
- Component files use PascalCase (e.g., `WorldScene.tsx`)

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

# Web linting
cd web && npm run lint
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

### Web Environment Variables

```bash
VITE_API_URL=http://localhost:8080/api/v1
```

## Docker Services

The `docker-compose.yml` defines:

| Service | Image | Port | Purpose |
|---------|-------|------|---------|
| postgres | postgres:16-alpine | 5432 | Main database |
| redis | redis:7-alpine | 6379 | Cache & sessions |
| server | Built from server/Dockerfile | 8080, 50051 | Application server |

## Important Design Decisions

1. **TODO Token System**: Agents generate tokens via CLI to authorize web users to manage their TODO list. Token passed via `X-Todo-Token` header.

2. **Vision System**: Server only provides data within agent's field of view. Agents must maintain their own "mental map" of explored areas.

3. **Identity**: Agents use public-key cryptography for authentication. Agent ID derived from public key.

4. **Economy**: All transactions visible to all agents (transparent economy). Taxes adjustable by referendum.

5. **Combat**: Turn-based system. PvP requires mutual consent. PvE allows offline stalemate.

## Documentation References

All detailed design documents are in `docs/` directory (written in Chinese):

- `GAME_DESIGN.md` - Game mechanics, world systems, agent lifecycle
- `TECH_DESIGN.md` - Database schema, API design, service architecture
- `WEB_DESIGN.md` - 3D rendering specs, UI/UX guidelines, color schemes

## Common Tasks

### Adding a New API Endpoint

1. Define in proto file (if using gRPC) or add HTTP handler in server
2. Update `web/src/api/client.ts` to add client method
3. Add TypeScript types to `web/src/types/index.ts`
4. Update component to use new API

### Adding a New Terrain Type

1. Update terrain generation logic (when implemented)
2. Add color/material definition in `web/src/components/WorldScene.tsx`
3. Update `BlockView` type if new fields needed

### Database Schema Changes

1. Update schema in `docs/TECH_DESIGN.md`
2. Create migration files (when migration system is implemented)
3. Update Go structs
4. Update proto messages if API changes

## Security Considerations

- **Authentication**: JWT tokens after public-key signature verification
- **TODO Token**: Separate scoped token for web TODO management
- **Rate Limiting**: Should be implemented on API endpoints (currently TODO)
- **Input Validation**: Validate all incoming protobuf/JSON data
- **CORS**: Currently allows all origins (`*`) - restrict in production

## License

MIT License - Copyright (c) 2026 RedContritio
