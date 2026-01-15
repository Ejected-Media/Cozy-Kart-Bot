This is the final piece of the connectivity puzzle: The WebSocket Hub.
In a standard web server, the server waits for a request, sends a reply, and then forgets the user exists. That doesn't work for a racing game. We need a persistent "pipe" that stays open so we can push updates (like "Pot is now $50!") instantly.
We will use the industry-standard gorilla/websocket library.
The Architecture: "The Fan-Out"
This pattern is called Fan-Out. A single update comes from the Game Engine, enters the Hub, and is "fanned out" to hundreds of connected phones simultaneously.
pkg/api/websocket.go
This file handles the connection upgrade and manages the pool of active users.
package api

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yourname/cozy-kart/pkg/game"
)

// 1. The Upgrader
// This helper converts a standard HTTP GET request into a WebSocket connection.
// CheckOrigin is set to true to allow connections from the Android Emulator/Phone.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true 
	},
}

// 2. The API Handler
// This struct holds the Game State and the list of connected phones (the Hub).
type Handler struct {
	Manager *game.RaceManager
	Clients map[*websocket.Conn]bool // The "Hub" (Set of active connections)
	mu      sync.Mutex               // Protects the Clients map from race conditions
}

// NewHandler creates the API controller.
func NewHandler(m *game.RaceManager) *Handler {
	h := &Handler{
		Manager: m,
		Clients: make(map[*websocket.Conn]bool),
	}
	
	// Start the background routine that watches for game updates
	go h.BroadcastLoop()
	
	return h
}

// 3. The Connection Endpoint
// "GET /ws" -> Android calls this to connect.
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to a WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ Upgrade Error: %v", err)
		return
	}

	// Register the new client
	h.mu.Lock()
	h.Clients[conn] = true
	h.mu.Unlock()
	
	log.Println("📱 New Client Connected")

	// Send them the initial state immediately so the screen isn't empty
	initialState := h.Manager.GetState() // helper method in RaceManager
	conn.WriteJSON(initialState)

	// Keep the connection alive until the client disconnects
	// We read messages here (even if we don't expect any from the client)
	// just to detect when the connection closes.
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			h.removeClient(conn)
			break
		}
	}
}

// Helper to clean up when a phone disconnects
func (h *Handler) removeClient(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.Clients, conn)
	h.mu.Unlock()
	conn.Close()
	log.Println("🔌 Client Disconnected")
}

// 4. The Broadcast Loop
// This runs forever in the background. It checks the Race Manager for changes
// and pushes them to everyone.
func (h *Handler) BroadcastLoop() {
	ticker := time.NewTicker(500 * time.Millisecond) // Update 2x per second
	defer ticker.Stop()

	for {
		<-ticker.C // Wait for the tick
		
		// Get the latest snapshot of the race
		// (Pot amount, Time remaining, Current phase)
		payload := h.Manager.GetState()

		// Send it to everyone
		h.mu.Lock()
		for conn := range h.Clients {
			err := conn.WriteJSON(payload)
			if err != nil {
				// If we can't write, assume they are gone and close it.
				log.Printf("⚠️ Write Error: %v", err)
				conn.Close()
				delete(h.Clients, conn)
			}
		}
		h.mu.Unlock()
	}
}

Key Technical Details
 * Thread Safety (sync.Mutex):
   * Since hundreds of phones connect/disconnect at random times (on different threads), and the BroadcastLoop runs on its own thread, we must lock the Clients map before touching it. If we don't, the server will panic and crash.
 * The Heartbeat (BroadcastLoop):
   * Instead of making the Game Engine complex by having it "push" updates, we use a simple Ticker.
   * Every 500ms, the API asks the Game: "What's the status?" and sends that JSON to everyone.
   * This is perfect for a racing dashboard (updates 2x/second is plenty for money/time counters).
 * JSON Payload:
   * The conn.WriteJSON(payload) line automatically converts your Go Struct (Money, Time, etc.) into a JSON string that the Kotlin Android app (using GSON/Moshi) can read instantly.
Next Steps
You now have the Complete Loop:
 * Discord sends a command (/host bonus 500).
 * Bot tells RaceManager (Pot += 500).
 * API Handler sees the change.
 * WebSocket pushes JSON to the Android Phone.
 * Phone Screen flashes Gold.
Would you like me to create a "Cheat Sheet" Summary Graphic of the entire project architecture (Files, Tech Stack, Data Flow) to save as a reference for this project?
