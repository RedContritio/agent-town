// Debug Server - 用于 godot-web 调试的混合服务器
//
// 架构:
//   client (HTTP) ←──→ debug_server:8081 ←──WebSocket──→ godot-web
//
// HTTP API (client 端):
//   POST /api/camera      - 设置相机参数
//   POST /api/preset      - 使用预设视角
//   POST /api/capture     - 请求截图
//   GET  /api/info        - 获取相机信息
//   GET  /health          - 健康检查
//
// WebSocket (godot-web 端):
//   ws://localhost:8081/godot

package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	addr     = flag.String("addr", ":8081", "Server address")
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

// Server 管理 godot WebSocket 连接和命令队列
type Server struct {
	mu sync.RWMutex

	godotConn  *websocket.Conn
	godotReady bool

	// 响应通道：godot 执行完命令后通过这里返回给 HTTP handler
	responseCh chan Response
}

// Response 通用响应
type Response struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func NewServer() *Server {
	return &Server{
		responseCh: make(chan Response, 10),
	}
}

// 注册 godot 连接 (只允许一个)
func (s *Server) registerGodot(conn *websocket.Conn) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.godotConn != nil {
		return false
	}
	s.godotConn = conn
	log.Println("[Server] Godot connected")
	return true
}

// 注销 godot 连接
func (s *Server) unregisterGodot() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.godotConn != nil {
		s.godotConn.Close()
		s.godotConn = nil
		s.godotReady = false
		log.Println("[Server] Godot disconnected")
	}
}

// 设置 godot ready 状态
func (s *Server) setGodotReady(ready bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.godotReady = ready
	if ready {
		log.Println("[Server] Godot is ready")
	}
}

// 检查 godot 是否 ready
func (s *Server) isGodotReady() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.godotReady
}

// 发送命令到 godot
func (s *Server) sendToGodot(cmd Response) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.godotConn == nil || !s.godotReady {
		log.Printf("[Server] Cannot send: conn=%v ready=%v", s.godotConn != nil, s.godotReady)
		return false
	}

	data, _ := json.Marshal(cmd)
	log.Printf("[Server] Sending to godot: %s", string(data))
	if err := s.godotConn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("[Server] Write to godot error: %v", err)
		return false
	}
	log.Printf("[Server] Message sent successfully")
	return true
}

// WebSocket handler for godot-web
func (s *Server) handleGodot(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[Server] Godot upgrade error: %v", err)
		return
	}
	defer conn.Close()

	if !s.registerGodot(conn) {
		conn.WriteJSON(Response{Type: "error", Payload: "server_busy"})
		return
	}
	defer s.unregisterGodot()

	// 发送连接确认
	conn.WriteJSON(Response{Type: "connected", Payload: map[string]string{"role": "godot"}})

	// 处理 godot 消息
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[Server] Godot read error: %v", err)
			break
		}

		var msg Response
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		log.Printf("[Server] From godot: %s", msg.Type)

		switch msg.Type {
		case "ready":
			s.setGodotReady(true)
		case "camera_info", "screenshot", "ack", "error":
			// 转发到 HTTP handler
			log.Printf("[Server] Forwarding %s to responseCh", msg.Type)
			select {
			case s.responseCh <- msg:
				log.Printf("[Server] Sent to responseCh (type=%s)", msg.Type)
			default:
				log.Printf("[Server] responseCh full or blocked, dropping message")
			}
		}
	}
}

// HTTP handlers for client
func (s *Server) handleCamera(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !s.isGodotReady() {
		http.Error(w, `{"error":"godot_not_ready"}`, http.StatusServiceUnavailable)
		return
	}

	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.sendToGodot(Response{Type: "set_camera", Payload: payload})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handlePreset(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !s.isGodotReady() {
		http.Error(w, `{"error":"godot_not_ready"}`, http.StatusServiceUnavailable)
		return
	}

	var payload map[string]string
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.sendToGodot(Response{Type: "set_preset", Payload: payload})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleCapture(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !s.isGodotReady() {
		http.Error(w, `{"error":"godot_not_ready"}`, http.StatusServiceUnavailable)
		return
	}

	var payload map[string]int
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		payload = map[string]int{"width": 1280, "height": 720}
	}

	// 清空之前的响应
	select {
	case <-s.responseCh:
	default:
	}

	// 发送截图命令
	log.Printf("[Server] Sending capture_screenshot to godot")
	s.sendToGodot(Response{Type: "capture_screenshot", Payload: payload})

	// 等待响应 (带超时)
	timeout := make(chan bool, 1)
	go func() { time.Sleep(10 * time.Second); timeout <- true }()

	select {
	case resp := <-s.responseCh:
		log.Printf("[Server] Got response from responseCh: %s", resp.Type)
		if resp.Type == "screenshot" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp.Payload)
		} else if resp.Type == "error" {
			http.Error(w, resp.Payload.(string), http.StatusInternalServerError)
		} else {
			// 不是截图响应，继续等待
			select {
			case resp = <-s.responseCh:
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp.Payload)
			case <-timeout:
				http.Error(w, `{"error":"timeout"}`, http.StatusGatewayTimeout)
			}
		}
	case <-timeout:
		http.Error(w, `{"error":"timeout"}`, http.StatusGatewayTimeout)
	}
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !s.isGodotReady() {
		http.Error(w, `{"error":"godot_not_ready"}`, http.StatusServiceUnavailable)
		return
	}

	// 清空之前的响应
	select {
	case <-s.responseCh:
	default:
	}

	// 发送获取信息命令
	s.sendToGodot(Response{Type: "get_camera_info"})

	// 等待响应
	timeout := make(chan bool, 1)
	go func() { time.Sleep(5 * time.Second); timeout <- true }()

	select {
	case resp := <-s.responseCh:
		if resp.Type == "camera_info" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp.Payload)
		} else {
			http.Error(w, `{"error":"unexpected_response"}`, http.StatusInternalServerError)
		}
	case <-timeout:
		http.Error(w, `{"error":"timeout"}`, http.StatusGatewayTimeout)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	godotOK := s.godotReady
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"godot_ready": godotOK,
	})
}

func main() {
	flag.Parse()

	server := NewServer()

	// WebSocket endpoint for godot
	http.HandleFunc("/godot", server.handleGodot)

	// HTTP API endpoints for client
	http.HandleFunc("/api/camera", server.handleCamera)
	http.HandleFunc("/api/preset", server.handlePreset)
	http.HandleFunc("/api/capture", server.handleCapture)
	http.HandleFunc("/api/info", server.handleInfo)
	http.HandleFunc("/health", server.handleHealth)

	log.Printf("[Server] Starting on %s", *addr)
	log.Printf("[Server] HTTP API: http://%s/api/{camera,preset,capture,info}", *addr)
	log.Printf("[Server] WebSocket: ws://%s/godot", *addr)

	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal(err)
	}
}
