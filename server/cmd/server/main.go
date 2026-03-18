package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

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

type BuildingView struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	OwnerID  string   `json:"ownerId"`
	Anchor   Position `json:"anchor"`
	Width    int      `json:"width"`
	Depth    int      `json:"depth"`
	Height   int      `json:"height"`
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
		ID:      building.ID,
		Name:    building.Name,
		Type:    building.Type,
		OwnerID: building.OwnerID,
		Anchor:  Position{X: building.Anchor.X, Y: building.Anchor.Y, Z: building.Anchor.Z},
		Width:   building.Width,
		Depth:   building.Depth,
		Height:  building.Height,
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
	w := world.GetWorld()
	if w == nil {
		http.Error(w, "World not initialized", http.StatusInternalServerError)
		return
	}

	info := WorldInfo{
		ID:            w.ID,
		Name:          "Agent Town",
		Seed:          fmt.Sprintf("%d", w.Seed),
		TimeSpeed:     5,
		CurrentTime:   time.Now().Format(time.RFC3339),
		AgentCount:    int64(len(w.InitialAgents)),
		BuildingCount: int64(len(w.InitialBuildings)),
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

func getWorldMap(w http.ResponseWriter, r *http.Request) {
	centerX := 0
	centerY := 0
	radius := 20

	// Parse query params
	if x := r.URL.Query().Get("x"); x != "" {
		fmt.Sscanf(x, "%d", &centerX)
	}
	if y := r.URL.Query().Get("y"); y != "" {
		fmt.Sscanf(y, "%d", &centerY)
	}
	if r := r.URL.Query().Get("radius"); r != "" {
		fmt.Sscanf(r, "%d", &radius)
	}

	wld := world.GetWorld()
	if wld == nil {
		http.Error(w, "World not initialized", http.StatusInternalServerError)
		return
	}

	// Get blocks from chunk manager
	blocks := []BlockView{}
	wldBlocks, err := wld.GetBlocksInRadius(centerX, centerY, radius)
	if err == nil {
		for _, b := range wldBlocks {
			blocks = append(blocks, convertBlock(b))
		}
	}

	// Get agents
	agents := []AgentView{}
	for _, a := range wld.InitialAgents {
		// Only return agents within radius
		dx := a.Position.X - centerX
		dy := a.Position.Y - centerY
		if dx*dx+dy*dy <= radius*radius {
			agents = append(agents, convertAgent(a))
		}
	}

	// Get buildings
	buildings := []BuildingView{}
	for _, b := range wld.InitialBuildings {
		// Only return buildings within radius
		dx := b.Anchor.X - centerX
		dy := b.Anchor.Y - centerY
		if dx*dx+dy*dy <= radius*radius {
			buildings = append(buildings, convertBuilding(b))
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"center":    Position{X: centerX, Y: centerY, Z: 0},
		"radius":    radius,
		"blocks":    blocks,
		"agents":    agents,
		"buildings": buildings,
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
	// Initialize world
	seed := int64(123456789)
	if err := world.InitWorld(seed, nil); err != nil {
		log.Fatalf("Failed to initialize world: %v", err)
	}

	// API routes
	http.HandleFunc("/api/v1/world/info", enableCORS(getWorldInfo))
	http.HandleFunc("/api/v1/world/time", enableCORS(getWorldTime))
	http.HandleFunc("/api/v1/world/map", enableCORS(getWorldMap))
	http.HandleFunc("/api/v1/agents", enableCORS(getAgents))
	http.HandleFunc("/api/v1/agents/", enableCORS(func(w http.ResponseWriter, r *http.Request) {
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

	// Static files for web
	fs := http.FileServer(http.Dir("./web/dist"))
	http.Handle("/", fs)

	port := "8080"
	fmt.Printf("Server starting on http://localhost:%s\n", port)
	fmt.Println("API endpoints:")
	fmt.Println("  GET /api/v1/world/info")
	fmt.Println("  GET /api/v1/world/time")
	fmt.Println("  GET /api/v1/agents")
	fmt.Println("  GET /api/v1/agents/{id}")
	fmt.Println("  GET /api/v1/agents/{id}/status")
	fmt.Println("  GET /api/v1/agents/{id}/visible-area")
	fmt.Println("  GET /api/v1/agents/{id}/todos")
	fmt.Println("  GET /api/v1/agents/{id}/skills")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
