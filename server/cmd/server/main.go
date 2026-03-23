package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/RedContritio/agent-town/server/internal/api"
	"github.com/RedContritio/agent-town/server/internal/db"
	"github.com/RedContritio/agent-town/server/internal/engine"
	"github.com/RedContritio/agent-town/server/internal/middleware"
	"github.com/RedContritio/agent-town/server/internal/repository"
	"github.com/RedContritio/agent-town/server/internal/service"
	"github.com/RedContritio/agent-town/server/internal/world"
)

// API response structures
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

type AgentView struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Position   Position `json:"position"`
	Facing     int      `json:"facing"`
	HP         int      `json:"hp"`
	MaxHP      int      `json:"maxHp"`
	Stamina    int      `json:"stamina"`
	MaxStamina int      `json:"maxStamina"`
	Hunger     int      `json:"hunger"`
	MaxHunger  int      `json:"maxHunger"`
	Balance    int      `json:"balance"`
	IsOnline   bool     `json:"isOnline"`
	InBattle   bool     `json:"inBattle"`
}

type BlockView struct {
	Position       Position `json:"position"`
	Height         int      `json:"height"`
	TerrainType    string   `json:"terrainType"`
	ResourceType   string   `json:"resourceType,omitempty"`
	ResourceAmount int      `json:"resourceAmount,omitempty"`
}

type ChunkView struct {
	X      int         `json:"x"`
	Y      int         `json:"y"`
	Blocks []BlockView `json:"blocks"`
}

type BuildingView struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	DisplayName string   `json:"displayName"`
	Type        string   `json:"type"`
	OwnerID     string   `json:"ownerId"`
	Anchor      Position `json:"anchor"`
	Width       int      `json:"width"`
	Depth       int      `json:"depth"`
	Height      int      `json:"height"`
}

type RoadView struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type VisibleArea struct {
	AgentID   string         `json:"agentId"`
	Center    Position       `json:"center"`
	Radius    int            `json:"radius"`
	Blocks    []BlockView    `json:"blocks"`
	Agents    []AgentView    `json:"agents"`
	Buildings []BuildingView `json:"buildings"`
}

type WorldInfo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Seed          string `json:"seed"`
	TimeSpeed     int    `json:"timeSpeed"`
	CurrentTime   string `json:"currentTime"`
	AgentCount    int64  `json:"agentCount"`
	BuildingCount int64  `json:"buildingCount"`
}

type WorldTime struct {
	Timestamp string `json:"timestamp"`
	Year      int    `json:"year"`
	Month     int    `json:"month"`
	Day       int    `json:"day"`
	Hour      int    `json:"hour"`
	Minute    int    `json:"minute"`
	Season    string `json:"season"`
	IsDaytime bool   `json:"isDaytime"`
}

type Todo struct {
	ID           string `json:"id"`
	Content      string `json:"content"`
	Status       string `json:"status"`
	Priority     int    `json:"priority"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
	RejectReason string `json:"rejectReason,omitempty"`
}

type Skill struct {
	Type      string `json:"type"`
	Level     int    `json:"level"`
	Exp       int    `json:"exp"`
	ExpToNext int    `json:"expToNext"`
}

type Event struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Data      string `json:"data"`
	CreatedAt string `json:"createdAt"`
}

// Convert internal world types to API types
func convertAgent(agent world.Agent) AgentView {
	return AgentView{
		ID:         agent.ID,
		Name:       agent.Name,
		Position:   Position{X: agent.Position.X, Y: agent.Position.Y, Z: agent.Position.Z},
		Facing:     agent.Facing,
		HP:         agent.HP,
		MaxHP:      agent.MaxHP,
		Stamina:    agent.Stamina,
		MaxStamina: agent.MaxStamina,
		Hunger:     agent.Hunger,
		MaxHunger:  agent.MaxHunger,
		Balance:    agent.Balance,
		IsOnline:   agent.IsOnline,
		InBattle:   agent.InBattle,
	}
}

func convertBuilding(building world.Building) BuildingView {
	return BuildingView{
		ID:          building.ID,
		Name:        building.Name,
		DisplayName: building.DisplayName,
		Type:        building.Type,
		OwnerID:     building.OwnerID,
		Anchor:      Position{X: building.Anchor.X, Y: building.Anchor.Y, Z: building.Anchor.Z},
		Width:       building.Width,
		Depth:       building.Depth,
		Height:      building.Height,
	}
}

func convertBlock(block world.Block) BlockView {
	return BlockView{
		Position:       Position{X: block.X, Y: block.Y, Z: block.Z},
		Height:         block.Height,
		TerrainType:    block.TerrainType,
		ResourceType:   block.ResourceType,
		ResourceAmount: block.ResourceAmount,
	}
}

// CORS middleware
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Todo-Token")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// Handlers
func getWorldInfo(w http.ResponseWriter, r *http.Request) {
	gameWorld := world.GetWorld()
	if gameWorld == nil {
		http.Error(w, "World not initialized", http.StatusInternalServerError)
		return
	}

	info := WorldInfo{
		ID:            gameWorld.ID,
		Name:          "Agent Town",
		Seed:          fmt.Sprintf("%d", gameWorld.Seed),
		TimeSpeed:     5,
		CurrentTime:   time.Now().Format(time.RFC3339),
		AgentCount:    int64(len(gameWorld.InitialAgents)),
		BuildingCount: int64(len(gameWorld.InitialBuildings)),
	}
	json.NewEncoder(w).Encode(info)
}

func getWorldTime(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	worldTime := WorldTime{
		Timestamp: now.Format(time.RFC3339),
		Year:      1,
		Month:     3,
		Day:       15,
		Hour:      now.Hour(),
		Minute:    now.Minute(),
		Season:    "spring",
		IsDaytime: now.Hour() >= 6 && now.Hour() < 18,
	}
	json.NewEncoder(w).Encode(worldTime)
}

const (
	MaxChunksPerRequest = 250  // 硬上限：最多一次获取 250 个 chunks
	ChunkSize           = 32   // 每个 chunk 32x32 方块
)

func getWorldMap(w http.ResponseWriter, r *http.Request) {
	// 新的 chunk-based 参数
	cx := 0  // 中心 chunk X
	cy := 0  // 中心 chunk Y
	cr := 2  // chunk 半径（默认 2，即 5x5=25 个 chunks）
	
	// 向后兼容：也支持世界坐标参数
	wx := 0  // 世界坐标 X
	wy := 0  // 世界坐标 Y
	useChunkMode := false

	// Parse query params - 优先检查 chunk 模式
	if v := r.URL.Query().Get("cx"); v != "" {
		fmt.Sscanf(v, "%d", &cx)
		useChunkMode = true
	}
	if v := r.URL.Query().Get("cy"); v != "" {
		fmt.Sscanf(v, "%d", &cy)
		useChunkMode = true
	}
	if v := r.URL.Query().Get("cr"); v != "" {
		fmt.Sscanf(v, "%d", &cr)
	}

	wld := world.GetWorld()
	if wld == nil {
		http.Error(w, "World not initialized", http.StatusInternalServerError)
		return
	}

	// 如果使用的是旧的世界坐标参数，转换为 chunk 坐标
	if !useChunkMode {
		if v := r.URL.Query().Get("x"); v != "" {
			fmt.Sscanf(v, "%d", &wx)
		}
		if v := r.URL.Query().Get("y"); v != "" {
			fmt.Sscanf(v, "%d", &wy)
		}
		// 将世界坐标转换为中心 chunk 坐标
		cx, cy, _, _ = wld.ChunkManager.WorldToChunk(wx, wy)
		
		// 如果有 radius 参数，转换为 chunk 半径
		if r := r.URL.Query().Get("radius"); r != "" {
			radius := 0
			fmt.Sscanf(r, "%d", &radius)
			// 计算需要多少 chunk 半径才能覆盖该世界坐标半径
			cr = (radius / ChunkSize) + 1
		}
	}

	// 限制 chunk 半径，确保不超过 MaxChunksPerRequest
	// 最大 chunks 数 = (2*cr+1)^2 <= MaxChunksPerRequest
	// 解得 cr <= (sqrt(MaxChunksPerRequest)-1)/2
	maxChunkRadius := int((math.Sqrt(MaxChunksPerRequest) - 1) / 2)
	if cr > maxChunkRadius {
		cr = maxChunkRadius
	}

	// Get chunks in chunk radius
	chunks := []ChunkView{}
	chunkCount := 0
	
	for dcx := -cr; dcx <= cr; dcx++ {
		for dcy := -cr; dcy <= cr; dcy++ {
			// 硬上限检查
			if chunkCount >= MaxChunksPerRequest {
				break
			}
			
			chunkX, chunkY := cx+dcx, cy+dcy
			chunk, err := wld.ChunkManager.GetChunk(chunkX, chunkY)
			if err == nil {
				chunkView := ChunkView{
					X: chunkX,
					Y: chunkY,
				}
				for _, b := range chunk.Blocks {
					chunkView.Blocks = append(chunkView.Blocks, convertBlock(b))
				}
				chunks = append(chunks, chunkView)
				chunkCount++
			}
		}
		if chunkCount >= MaxChunksPerRequest {
			break
		}
	}

	// 计算世界坐标边界用于过滤 agents/buildings
	minWorldX := (cx - cr) * ChunkSize
	maxWorldX := (cx + cr + 1) * ChunkSize
	minWorldY := (cy - cr) * ChunkSize
	maxWorldY := (cy + cr + 1) * ChunkSize

	// Get agents within chunk bounds
	agents := []AgentView{}
	for _, a := range wld.InitialAgents {
		if a.Position.X >= minWorldX && a.Position.X < maxWorldX &&
		   a.Position.Y >= minWorldY && a.Position.Y < maxWorldY {
			agents = append(agents, convertAgent(a))
		}
	}

	// Get buildings within chunk bounds
	buildings := []BuildingView{}
	for _, b := range wld.InitialBuildings {
		if b.Anchor.X >= minWorldX && b.Anchor.X < maxWorldX &&
		   b.Anchor.Y >= minWorldY && b.Anchor.Y < maxWorldY {
			buildings = append(buildings, convertBuilding(b))
		}
	}

	// Get roads within chunk bounds
	roads := []RoadView{}
	for _, r := range wld.InitialRoads {
		if r.X >= minWorldX && r.X < maxWorldX &&
		   r.Y >= minWorldY && r.Y < maxWorldY {
			roads = append(roads, RoadView{X: r.X, Y: r.Y})
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"centerChunk": Position{X: cx, Y: cy, Z: 0},
		"chunkRadius": cr,
		"chunks":      chunks,
		"agents":      agents,
		"buildings":   buildings,
		"roads":       roads,
	})
}

func getAgents(w http.ResponseWriter, r *http.Request) {
	wld := world.GetWorld()
	if wld == nil {
		http.Error(w, "World not initialized", http.StatusInternalServerError)
		return
	}

	agents := []AgentView{}
	for _, a := range wld.InitialAgents {
		agents = append(agents, convertAgent(a))
	}
	json.NewEncoder(w).Encode(agents)
}

func getAgent(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Path[len("/api/v1/agents/"):]
	wld := world.GetWorld()
	if wld == nil {
		http.Error(w, "World not initialized", http.StatusInternalServerError)
		return
	}

	for _, agent := range wld.InitialAgents {
		if agent.ID == agentID {
			json.NewEncoder(w).Encode(convertAgent(agent))
			return
		}
	}
	http.NotFound(w, r)
}

func getAgentStatus(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Path[len("/api/v1/agents/"):len(r.URL.Path)-len("/status")]
	wld := world.GetWorld()
	if wld == nil {
		http.Error(w, "World not initialized", http.StatusInternalServerError)
		return
	}

	for _, agent := range wld.InitialAgents {
		if agent.ID == agentID {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"agentId":       agent.ID,
				"worldTime":     time.Now().Format(time.RFC3339),
				"stamina":       agent.Stamina,
				"hunger":        agent.Hunger,
				"inBattle":      agent.InBattle,
				"pendingEvents": []Event{},
			})
			return
		}
	}
	http.NotFound(w, r)
}

func getVisibleArea(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Path[len("/api/v1/agents/"):len(r.URL.Path)-len("/visible-area")]
	wld := world.GetWorld()
	if wld == nil {
		http.Error(w, "World not initialized", http.StatusInternalServerError)
		return
	}

	// Find agent position
	var agentPos Position
	for _, agent := range wld.InitialAgents {
		if agent.ID == agentID {
			agentPos = Position{X: agent.Position.X, Y: agent.Position.Y, Z: agent.Position.Z}
			break
		}
	}

	radius := 10

	// Get blocks
	blocks := []BlockView{}
	wldBlocks, err := wld.GetBlocksInRadius(agentPos.X, agentPos.Y, radius)
	if err == nil {
		for _, b := range wldBlocks {
			blocks = append(blocks, convertBlock(b))
		}
	}

	// Get agents in range
	agents := []AgentView{}
	for _, a := range wld.InitialAgents {
		dx := a.Position.X - agentPos.X
		dy := a.Position.Y - agentPos.Y
		if dx*dx+dy*dy <= radius*radius {
			agents = append(agents, convertAgent(a))
		}
	}

	// Get buildings in range
	buildings := []BuildingView{}
	for _, b := range wld.InitialBuildings {
		dx := b.Anchor.X - agentPos.X
		dy := b.Anchor.Y - agentPos.Y
		if dx*dx+dy*dy <= radius*radius {
			buildings = append(buildings, convertBuilding(b))
		}
	}

	area := VisibleArea{
		AgentID:   agentID,
		Center:    agentPos,
		Radius:    radius,
		Blocks:    blocks,
		Agents:    agents,
		Buildings: buildings,
	}
	json.NewEncoder(w).Encode(area)
}

// Mock todos and skills (to be implemented later)
var mockTodos = []Todo{
	{
		ID:        "todo-001",
		Content:   "Gather wood from the forest",
		Status:    "pending",
		Priority:  5,
		CreatedAt: time.Now().Add(-time.Hour).Format(time.RFC3339),
		UpdatedAt: time.Now().Add(-time.Hour).Format(time.RFC3339),
	},
	{
		ID:        "todo-002",
		Content:   "Build a storage chest",
		Status:    "planning",
		Priority:  7,
		CreatedAt: time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
		UpdatedAt: time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
	},
}

var mockSkills = []Skill{
	{Type: "farming", Level: 3, Exp: 150, ExpToNext: 200},
	{Type: "mining", Level: 2, Exp: 80, ExpToNext: 150},
	{Type: "building", Level: 4, Exp: 300, ExpToNext: 400},
}

func getTodos(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agentId": "agent-001",
		"todos":   mockTodos,
	})
}

func getSkills(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agentId": "agent-001",
		"skills":  mockSkills,
	})
}

func main() {
	// Initialize database
	dbConfig := db.DefaultConfig()
	database, err := db.Init(dbConfig)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Println("Database initialized")
	
	// Ensure engine stops on exit
	defer engine.GetManager().Stop()

	// Initialize world
	seed := int64(123456789)
	if err := world.InitWorld(seed, database); err != nil {
		log.Fatalf("Failed to initialize world: %v", err)
	}

	// Initialize repositories
	agentRepo := repository.NewAgentRepository(database)
	tokenRepo := repository.NewTokenRepository(database)
	skillsRepo := repository.NewSkillsRepository(database)
	invRepo := repository.NewInventoryRepository(database)
	taskRepo := repository.NewTaskRepository(database)

	// Initialize services
	authService := service.NewAuthService(agentRepo, tokenRepo)
	agentService := service.NewAgentService(agentRepo, skillsRepo, invRepo)
	craftService := service.NewCraftService(invRepo, skillsRepo)
	taskService := service.NewTaskService(taskRepo, agentRepo, craftService)
	invService := service.NewInventoryService(invRepo)

	// Initialize and start task engine
	worldRepo := repository.NewWorldRepository(database)
	engineManager := engine.GetManager()
	engineManager.Init(agentRepo, taskRepo, worldRepo, invRepo)
	engineManager.Start()
	log.Println("Task engine started")
	
	// Set craft handler after engine is initialized (avoid circular dependency)
	engineManager.SetCraftHandler(craftService)

	// Initialize vision service
	gameWorld := world.GetWorld()
	visionService := service.NewVisionService(worldRepo, gameWorld.ChunkManager)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Initialize handlers
	authHandler := api.NewAuthHandler(authService)
	agentHandler := api.NewAgentHandler(agentService, authMiddleware)
	taskHandler := api.NewTaskHandler(taskService, agentService, authMiddleware)
	visionHandler := api.NewVisionHandler(visionService, authMiddleware)
	invHandler := api.NewInventoryHandler(invService, authMiddleware)
	commandHandler := api.NewCommandHandler()

	// Create mux for routing
	mux := http.NewServeMux()

	// Auth routes
	authHandler.RegisterRoutes(mux)

	// Agent routes (me)
	agentHandler.RegisterRoutes(mux)

	// Task routes (与 CLI 对齐)
	taskHandler.RegisterRoutes(mux)

	// Vision routes (与 CLI 对齐)
	visionHandler.RegisterRoutes(mux)

	// Inventory routes (与 CLI 对齐)
	invHandler.RegisterRoutes(mux)

	// Command routes
	commandHandler.RegisterRoutes(mux)

	// Existing API routes
	mux.HandleFunc("/api/v1/world/info", enableCORS(getWorldInfo))
	mux.HandleFunc("/api/v1/world/time", enableCORS(getWorldTime))
	mux.HandleFunc("/api/v1/world/map", enableCORS(getWorldMap))
	mux.HandleFunc("/api/v1/agents", enableCORS(getAgents))
	mux.HandleFunc("/api/v1/agents/", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case len(path) > len("/api/v1/agents/") && path[len(path)-7:] == "/status":
			getAgentStatus(w, r)
		case len(path) > len("/api/v1/agents/") && path[len(path)-13:] == "/visible-area":
			getVisibleArea(w, r)
		case len(path) > len("/api/v1/agents/") && path[len(path)-6:] == "/todos":
			getTodos(w, r)
		case len(path) > len("/api/v1/agents/") && path[len(path)-7:] == "/skills":
			getSkills(w, r)
		default:
			getAgent(w, r)
		}
	}))

	// Determine web directory: if ./web exists use it, else use executable's directory
	webDir := "./web"
	if _, err := os.Stat(webDir); os.IsNotExist(err) {
		// ./web not found, try relative to executable
		ex, err := os.Executable()
		if err != nil {
			log.Fatalf("Failed to get executable path: %v", err)
		}
		exPath := filepath.Dir(ex)
		webDir = filepath.Join(exPath, "web")
	}
	
	log.Printf("Serving static files from: %s", webDir)
	
	// Static files for web with proper MIME types and COOP/COEP headers for Godot
	fs := http.FileServer(http.Dir(webDir))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Set correct MIME types for Godot web export files
		if strings.HasSuffix(r.URL.Path, ".js") {
			w.Header().Set("Content-Type", "application/javascript")
		} else if strings.HasSuffix(r.URL.Path, ".wasm") {
			w.Header().Set("Content-Type", "application/wasm")
		} else if strings.HasSuffix(r.URL.Path, ".pck") {
			w.Header().Set("Content-Type", "application/octet-stream")
		} else if strings.HasSuffix(r.URL.Path, ".html") {
			w.Header().Set("Content-Type", "text/html")
		}
		
		// Cross-Origin Isolation headers required for Godot Web (SharedArrayBuffer)
		w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		
		// Add cache control headers to prevent caching
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		fs.ServeHTTP(w, r)
	})

	port := "8080"
	fmt.Printf("Server starting on http://localhost:%s\n", port)
	fmt.Println("API endpoints:")
	fmt.Println("  POST /api/v1/auth/register     (注册)")
	fmt.Println("  POST /api/v1/auth/challenge    (挑战)")
	fmt.Println("  POST /api/v1/auth/token        (获取 Token)")
	fmt.Println("  DELETE /api/v1/auth/logout     (登出)")
	fmt.Println("  GET /api/v1/commands           (命令列表)")
	fmt.Println("")
	fmt.Println("  # 与 CLI 命令对齐 (都需要认证)")
	fmt.Println("  GET /api/v1/status             (对应: status)")
	fmt.Println("  GET /api/v1/look               (对应: look)")
	fmt.Println("  GET /api/v1/scan               (对应: scan)")
	fmt.Println("  GET /api/v1/inventory          (对应: inventory)")
	fmt.Println("  GET /api/v1/stack              (对应: stack)")
	fmt.Println("  POST /api/v1/stack             (对应: move/harvest/craft/build)")
	fmt.Println("  DELETE /api/v1/stack/{id}      (对应: stack drop)")
	fmt.Println("  POST /api/v1/stack/{id}/pause  (对应: stack focus/pause)")
	fmt.Println("  POST /api/v1/stack/{id}/resume (对应: stack resume)")
	fmt.Println("  GET /api/v1/world/info")
	fmt.Println("  GET /api/v1/world/time")
	fmt.Println("  GET /api/v1/agents")
	fmt.Println("  GET /api/v1/agents/{id}")
	fmt.Println("  GET /api/v1/agents/{id}/status")
	fmt.Println("  GET /api/v1/agents/{id}/visible-area")
	fmt.Println("  GET /api/v1/agents/{id}/todos")
	fmt.Println("  GET /api/v1/agents/{id}/skills")
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
