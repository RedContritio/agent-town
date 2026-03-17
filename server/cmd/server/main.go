package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Mock data structures
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

type Agent struct {
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
	Agents    []Agent        `json:"agents"`
	Buildings []BuildingView `json:"buildings"`
}

type WorldInfo struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Seed           string `json:"seed"`
	TimeSpeed      int    `json:"timeSpeed"`
	CurrentTime    string `json:"currentTime"`
	AgentCount     int64  `json:"agentCount"`
	BuildingCount  int64  `json:"buildingCount"`
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
	Type       string `json:"type"`
	Level      int    `json:"level"`
	Exp        int    `json:"exp"`
	ExpToNext  int    `json:"expToNext"`
}

// Mock data
var mockAgents = []Agent{
	{
		ID:         "agent-001",
		Name:       "Alice",
		Position:   Position{X: 0, Y: 0, Z: 0},
		Facing:     0,
		HP:         100,
		MaxHP:      100,
		Stamina:    80,
		MaxStamina: 100,
		Hunger:     90,
		MaxHunger:  100,
		Balance:    1000,
		IsOnline:   true,
		InBattle:   false,
	},
	{
		ID:         "agent-002",
		Name:       "Bob",
		Position:   Position{X: 5, Y: 0, Z: 3},
		Facing:     1,
		HP:         100,
		MaxHP:      100,
		Stamina:    100,
		MaxStamina: 100,
		Hunger:     100,
		MaxHunger:  100,
		Balance:    500,
		IsOnline:   true,
		InBattle:   false,
	},
}

var mockBuildings = []BuildingView{
	{
		ID:      "building-001",
		Name:    "Town Hall",
		OwnerID: "government",
		Anchor:  Position{X: -5, Y: 0, Z: -5},
		Width:   4,
		Depth:   4,
		Height:  2,
	},
	{
		ID:      "building-002",
		Name:    "Shop",
		OwnerID: "government",
		Anchor:  Position{X: 5, Y: 0, Z: -5},
		Width:   3,
		Depth:   3,
		Height:  1,
	},
	{
		ID:      "building-003",
		Name:    "Alice's House",
		OwnerID: "agent-001",
		Anchor:  Position{X: 2, Y: 0, Z: 2},
		Width:   2,
		Depth:   2,
		Height:  1,
	},
}

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
	info := WorldInfo{
		ID:            "world-001",
		Name:          "Agent Town",
		Seed:          "123456789",
		TimeSpeed:     5,
		CurrentTime:   time.Now().Format(time.RFC3339),
		AgentCount:    int64(len(mockAgents)),
		BuildingCount: int64(len(mockBuildings)),
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
	radius := 10
	
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
	
	// Generate flat square blocks - no random hills
	blocks := []BlockView{}
	for x := centerX - radius; x <= centerX+radius; x++ {
		for y := centerY - radius; y <= centerY+radius; y++ {
			blocks = append(blocks, BlockView{
				Position:    Position{X: x, Y: y, Z: 0},
				Height:      0,
				TerrainType: "grass",
			})
		}
	}
	
	json.NewEncoder(w).Encode(map[string]interface{}{
		"center": Position{X: centerX, Y: centerY, Z: 0},
		"radius": radius,
		"blocks": blocks,
		"agents": mockAgents,
		"buildings": mockBuildings,
	})
}

func getAgents(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(mockAgents)
}

func getAgent(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Path[len("/api/v1/agents/"):]
	for _, agent := range mockAgents {
		if agent.ID == agentID {
			json.NewEncoder(w).Encode(agent)
			return
		}
	}
	http.NotFound(w, r)
}

func getAgentStatus(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Path[len("/api/v1/agents/"):len(r.URL.Path)-len("/status")]
	for _, agent := range mockAgents {
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

type Event struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Data      string `json:"data"`
	CreatedAt string `json:"createdAt"`
}

func getVisibleArea(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Path[len("/api/v1/agents/"):len(r.URL.Path)-len("/visible-area")]
	
	// Generate flat square blocks
	blocks := []BlockView{}
	for x := -5; x <= 5; x++ {
		for y := -5; y <= 5; y++ {
			blocks = append(blocks, BlockView{
				Position:    Position{X: x, Y: y, Z: 0},
				Height:      0,
				TerrainType: "grass",
			})
		}
	}
	
	area := VisibleArea{
		AgentID:   agentID,
		Center:    Position{X: 0, Y: 0, Z: 0},
		Radius:    10,
		Blocks:    blocks,
		Agents:    mockAgents,
		Buildings: mockBuildings,
	}
	json.NewEncoder(w).Encode(area)
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
