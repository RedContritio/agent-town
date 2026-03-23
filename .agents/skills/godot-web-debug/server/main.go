// Debug Server - 用于 Godot 调试的混合服务器
//
// 架构:
//   client (HTTP) ←──→ debug_server:8081 ←──WebSocket──→ Godot
//
// HTTP API (client 端):
//   POST /api/camera        - 设置相机参数（轨道模式）
//   POST /api/camera/direct - 直接设置相机位置（平视视角）
//   POST /api/preset        - 使用预设视角
//   POST /api/capture       - 请求截图
//   GET  /api/info          - 获取相机信息
//   GET  /health            - 健康检查
//
// WebSocket (Godot 端):
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

	if s.godotConn == nil {
		log.Printf("[Server] Cannot send: godotConn is nil")
		return false
	}
	if !s.godotReady {
		log.Printf("[Server] Cannot send: godot not ready")
		return false
	}

	data, _ := json.Marshal(cmd)
	log.Printf("[Server] Sending to godot: %s", string(data))
	if err := s.godotConn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("[Server] Write to godot error: %v", err)
		return false
	}
	log.Printf("[Server] Message sent successfully (type=%s)", cmd.Type)
	return true
}

// sendAndWait 发送命令并等待响应
func (s *Server) sendAndWait(cmd Response, timeout time.Duration) (Response, bool) {
	// 清空之前的响应，避免读到旧数据
	select {
	case <-s.responseCh:
		log.Printf("[Server] Cleared stale response from channel")
	default:
	}

	// 发送命令
	if !s.sendToGodot(cmd) {
		return Response{}, false
	}

	// 等待响应
	log.Printf("[Server] Waiting for response (timeout=%v)...", timeout)
	timeoutCh := time.After(timeout)

	for {
		select {
		case resp := <-s.responseCh:
			log.Printf("[Server] Received response: type=%s", resp.Type)
			// 如果是 ack 或 error，直接返回
			if resp.Type == "ack" || resp.Type == "error" {
				return resp, true
			}
			// 如果是其他类型（如截图），也返回
			if cmd.Type == "capture_screenshot" && resp.Type == "screenshot" {
				return resp, true
			}
			if cmd.Type == "get_camera_info" && resp.Type == "camera_info" {
				return resp, true
			}
			// 其他消息继续等待
			log.Printf("[Server] Unexpected response type=%s, continuing to wait...", resp.Type)
			continue
		case <-timeoutCh:
			log.Printf("[Server] Timeout waiting for response")
			return Response{}, false
		}
	}
}

// WebSocket handler for Godot
func (s *Server) handleGodot(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[Server] Godot upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// 设置 pong handler
	conn.SetPongHandler(func(appData string) error {
		log.Printf("[Server] Received pong from godot")
		return nil
	})

	if !s.registerGodot(conn) {
		conn.WriteJSON(Response{Type: "error", Payload: "server_busy"})
		return
	}
	defer s.unregisterGodot()

	// 发送连接确认
	conn.WriteJSON(Response{Type: "connected", Payload: map[string]string{"role": "godot"}})
	log.Printf("[Server] Godot WebSocket connected, waiting for ready message...")

	// 启动 ping ticker（保持连接活跃）
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// 在 goroutine 中发送 ping
	pingDone := make(chan bool)
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Printf("[Server] Ping error: %v", err)
					return
				}
				log.Printf("[Server] Sent ping to godot")
			case <-pingDone:
				return
			}
		}
	}()

	// 处理 godot 消息
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[Server] Godot read error: %v", err)
			close(pingDone)
			break
		}

		var msg Response
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("[Server] JSON parse error: %v", err)
			continue
		}

		log.Printf("[Server] From godot: type=%s", msg.Type)

		switch msg.Type {
		case "ready":
			s.setGodotReady(true)
			log.Printf("[Server] Godot is now ready")
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

	resp, ok := s.sendAndWait(Response{Type: "set_camera", Payload: payload}, 5*time.Second)
	if !ok {
		http.Error(w, `{"error":"timeout or send failed"}`, http.StatusGatewayTimeout)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if resp.Type == "error" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": resp.Payload.(string)})
	} else {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

func (s *Server) handleCameraDirect(w http.ResponseWriter, r *http.Request) {
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

	resp, ok := s.sendAndWait(Response{Type: "set_camera_direct", Payload: payload}, 5*time.Second)
	if !ok {
		http.Error(w, `{"error":"timeout or send failed"}`, http.StatusGatewayTimeout)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if resp.Type == "error" {
		w.WriteHeader(http.StatusBadRequest)
		var errorMsg string
		switch p := resp.Payload.(type) {
		case string:
			errorMsg = p
		case map[string]interface{}:
			if msg, ok := p["message"].(string); ok {
				errorMsg = msg
			} else {
				errorMsg = "unknown error"
			}
		default:
			errorMsg = "unknown error"
		}
		json.NewEncoder(w).Encode(map[string]string{"error": errorMsg})
	} else {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
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

	// 转换字段名: "name" -> "preset"
	godotPayload := make(map[string]string)
	if name, ok := payload["name"]; ok {
		godotPayload["preset"] = name
	}

	resp, ok := s.sendAndWait(Response{Type: "set_preset", Payload: godotPayload}, 5*time.Second)
	if !ok {
		http.Error(w, `{"error":"timeout or send failed"}`, http.StatusGatewayTimeout)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if resp.Type == "error" {
		w.WriteHeader(http.StatusBadRequest)
		// Payload could be a map or string
		var errorMsg string
		switch p := resp.Payload.(type) {
		case string:
			errorMsg = p
		case map[string]interface{}:
			if msg, ok := p["message"].(string); ok {
				errorMsg = msg
			} else {
				errorMsg = "unknown error"
			}
		default:
			errorMsg = "unknown error"
		}
		json.NewEncoder(w).Encode(map[string]string{"error": errorMsg})
	} else {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
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
	http.HandleFunc("/api/camera/direct", server.handleCameraDirect)
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
