package ui

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/ghchinoy/pandemonium/pkg/types"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool { return true },
}

// Server implements a WebSocket-based event broadcaster and static file server.
type Server struct {
	events  chan types.Event
	mu      sync.Mutex
	conns   map[*websocket.Conn]bool
	WebPath string
}

// NewServer initializes a new UI server.
func NewServer(eventSource chan types.Event, webPath string) *Server {
	return &Server{
		events:  eventSource,
		conns:   make(map[*websocket.Conn]bool),
		WebPath: webPath,
	}
}

// Start launches the HTTP and WebSocket server on the specified port.
func (s *Server) Start(port int) {
	http.HandleFunc("/ws", s.handleWS)
	// Serve static files from the configured directory
	http.Handle("/", http.FileServer(http.Dir(s.WebPath)))

	fmt.Printf("[Server] Starting X-Ray UI at http://localhost:%d\n", port)
	
	// Start event broadcaster
	go s.broadcast()

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
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
		data, _ := json.Marshal(e)
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