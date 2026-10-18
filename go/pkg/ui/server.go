package ui

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"

	"github.com/ghchinoy/pandemonium/pkg/engine"
	"github.com/ghchinoy/pandemonium/pkg/types"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool { return true },
}

// Server implements a WebSocket-based event broadcaster, REST control API, and static file server.
type Server struct {
	events     chan types.Event
	mu         sync.Mutex
	conns      map[*websocket.Conn]bool
	WebPath    string
	Controller *engine.StepController
	Manager    *engine.SimulationManager
}

// NewServer initializes a new UI server.
func NewServer(eventSource chan types.Event, webPath string) *Server {
	return &Server{
		events:  eventSource,
		conns:   make(map[*websocket.Conn]bool),
		WebPath: webPath,
	}
}

// Handler returns the HTTP handler mux with WebSocket, REST API, and static file routes.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWS)
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/control", s.handleControl)
	mux.HandleFunc("/api/scenarios", s.handleScenarios)
	mux.HandleFunc("/api/sessions", s.handleSessions)
	if s.WebPath != "" {
		mux.Handle("/", http.FileServer(http.Dir(s.WebPath)))
	}
	return mux
}

// Start launches the HTTP and WebSocket server on the specified port.
func (s *Server) Start(port int) {
	fmt.Printf("[Server] Starting X-Ray UI at http://localhost:%d\n", port)

	// Start event broadcaster
	go s.broadcast()

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), s.Handler()))
}

func (s *Server) activeController(sessionID string) *engine.StepController {
	if s.Manager != nil {
		if ctrl, ok := s.Manager.Controller(sessionID); ok {
			return ctrl
		}
	}
	return s.Controller
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	ctrl := s.activeController(sessionID)

	resp := map[string]any{
		"running": true,
		"paused":  false,
		"stopped": false,
	}
	if ctrl != nil {
		resp["running"] = ctrl.IsRunning()
		resp["paused"] = ctrl.IsPaused()
		resp["stopped"] = ctrl.ShouldStop()
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleControl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Action    string `json:"action"`
		SessionID string `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	ctrl := s.activeController(req.SessionID)
	if ctrl == nil {
		http.Error(w, "no active step controller", http.StatusNotFound)
		return
	}

	switch req.Action {
	case "play", "resume":
		ctrl.Play()
	case "pause":
		ctrl.Pause()
	case "step":
		ctrl.Step()
	case "stop":
		ctrl.Stop()
	default:
		http.Error(w, fmt.Sprintf("unknown control action %q", req.Action), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"action":  req.Action,
		"running": ctrl.IsRunning(),
		"paused":  ctrl.IsPaused(),
		"stopped": ctrl.ShouldStop(),
	})
}

func (s *Server) handleScenarios(w http.ResponseWriter, r *http.Request) {
	mgr := s.Manager
	if mgr == nil {
		mgr = engine.NewManager(nil, nil, nil, nil)
	}
	scenarios, err := mgr.ListScenarios(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("listing scenarios: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"scenarios": scenarios})
}

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	if s.Manager == nil {
		writeJSON(w, http.StatusOK, map[string]any{"sessions": []engine.SessionInfo{}})
		return
	}

	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{"sessions": s.Manager.ListSessions()})
	case http.MethodPost:
		var req struct {
			ID          string `json:"id"`
			Scenario    string `json:"scenario"`
			StartPaused bool   `json:"start_paused"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
			return
		}
		if req.ID == "" {
			req.ID = "default"
		}
		if req.Scenario == "" {
			req.Scenario = "pub.yaml"
		}
		ctrl, err := s.Manager.StartSession(r.Context(), req.ID, req.Scenario, req.StartPaused)
		if err != nil {
			http.Error(w, fmt.Sprintf("starting session: %v", err), http.StatusBadRequest)
			return
		}
		s.Controller = ctrl
		writeJSON(w, http.StatusCreated, map[string]any{
			"id":       req.ID,
			"scenario": req.Scenario,
			"running":  ctrl.IsRunning(),
			"paused":   ctrl.IsPaused(),
		})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}
	s.mu.Lock()
	s.conns[conn] = true
	s.mu.Unlock()
}

func (s *Server) broadcast() {
	for e := range s.events {
		data, err := json.Marshal(e)
		if err != nil {
			log.Printf("Error marshaling event: %v", err)
			continue
		}
		s.mu.Lock()
		for conn := range s.conns {
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				if cerr := conn.Close(); cerr != nil {
					log.Printf("Error closing websocket connection: %v", cerr)
				}
				delete(s.conns, conn)
			}
		}
		s.mu.Unlock()
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
