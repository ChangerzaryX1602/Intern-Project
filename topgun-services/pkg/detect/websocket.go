package detect

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"os"
	"sync"
	"topgun-services/pkg/models"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// WebSocket client structure
type Client struct {
	conn     *websocket.Conn
	cameraID uuid.UUID
	send     chan []byte
}

// WebSocket hub to manage clients
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan *BroadcastMessage
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
}

type BroadcastMessage struct {
	CameraID uuid.UUID
	Detect   *models.Detect
}

// Detection message with image
type DetectionMessage struct {
	ID        uint                       `json:"id"`
	CameraID  uuid.UUID                  `json:"camera_id"`
	Timestamp string                     `json:"timestamp"`
	Path      string                     `json:"path"`
	Objects   models.JSONRawMessageArray `json:"objects"`
	ImageData string                     `json:"image_data"` // Base64 encoded image
	MimeType  string                     `json:"mime_type"`  // image/jpeg, image/png, etc.
}

// Global hub instance
var hub *Hub

func init() {
	hub = &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *BroadcastMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
	go hub.run()
}

// Run hub to handle client connections and broadcasts
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			log.Printf("Client connected. Camera ID: %s. Total clients: %d", client.cameraID, len(h.clients))

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client disconnected. Camera ID: %s. Total clients: %d", client.cameraID, len(h.clients))
			}
			h.mutex.Unlock()

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				// Only send to clients subscribed to this camera
				if client.cameraID == message.CameraID {
					// Create message with image data
					detectionMsg := createDetectionMessage(message.Detect)
					select {
					case client.send <- mustMarshal(detectionMsg):
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// Create detection message with base64 encoded image
func createDetectionMessage(detect *models.Detect) *DetectionMessage {
	msg := &DetectionMessage{
		ID:        detect.ID,
		CameraID:  detect.CameraID,
		Timestamp: detect.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		Path:      detect.Path,
		Objects:   detect.Objects,
	}

	// Read and encode image file
	if detect.Path != "" {
		if imageData, err := os.ReadFile(detect.Path); err == nil {
			// Encode to base64
			msg.ImageData = base64.StdEncoding.EncodeToString(imageData)

			// Detect mime type from file extension
			msg.MimeType = getMimeType(detect.Path)
		} else {
			log.Printf("Failed to read image file %s: %v", detect.Path, err)
		}
	}

	return msg
}

// Get mime type from file extension
func getMimeType(path string) string {
	ext := ""
	for i := len(path) - 1; i >= 0 && path[i] != '/'; i-- {
		if path[i] == '.' {
			ext = path[i:]
			break
		}
	}

	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".bmp":
		return "image/bmp"
	case ".svg":
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}

// Broadcast detection to subscribed clients
func BroadcastDetection(detect *models.Detect) {
	if hub != nil {
		hub.broadcast <- &BroadcastMessage{
			CameraID: detect.CameraID,
			Detect:   detect,
		}
	}
}

// WebSocket handler
func (h *detectHandler) HandleWebSocket() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		var client *Client
		defer func() {
			if client != nil {
				hub.unregister <- client
				c.Close()
			}
		}()

		// Read initial message to get camera_id
		var subscribeMsg struct {
			CameraID string `json:"camera_id"`
		}

		if err := c.ReadJSON(&subscribeMsg); err != nil {
			log.Printf("Error reading subscribe message: %v", err)
			return
		}

		// Parse camera ID
		cameraID, err := uuid.Parse(subscribeMsg.CameraID)
		if err != nil {
			log.Printf("Invalid camera ID: %v", err)
			c.WriteJSON(fiber.Map{
				"error": "Invalid camera_id format",
			})
			return
		}

		// Create client
		client = &Client{
			conn:     c,
			cameraID: cameraID,
			send:     make(chan []byte, 256),
		}

		// Register client
		hub.register <- client

		// Send confirmation
		c.WriteJSON(fiber.Map{
			"status":    "subscribed",
			"camera_id": cameraID.String(),
		})

		// Start goroutine to write messages
		go func() {
			for message := range client.send {
				if err := c.WriteMessage(websocket.TextMessage, message); err != nil {
					log.Printf("Error writing message: %v", err)
					return
				}
			}
		}()

		// Keep connection alive and handle ping/pong
		for {
			messageType, _, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("Unexpected close error: %v", err)
				}
				break
			}

			// Handle ping
			if messageType == websocket.PingMessage {
				c.WriteMessage(websocket.PongMessage, nil)
			}
		}
	})
}

// Helper function to marshal JSON
func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return []byte("{}")
	}
	return data
}

// WebSocket upgrade middleware
func WebSocketUpgrade() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}
}
