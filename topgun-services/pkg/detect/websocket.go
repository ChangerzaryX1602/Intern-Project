package detect

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
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

// Video frame message from Python
type VideoFrameMessage struct {
	Frame       string  `json:"frame"`        // Base64 encoded JPEG
	Timestamp   float64 `json:"timestamp"`    // Unix timestamp
	FrameNumber int     `json:"frame_number"` // Frame sequence number
	Detections  int     `json:"detections"`   // Number of objects detected
	Width       int     `json:"width"`        // Frame width
	Height      int     `json:"height"`       // Frame height
	Model       string  `json:"model"`        // Model name used
}

// RaspberryPI MQTT Detection Data
type RaspberryPIDetection struct {
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	W          float64 `json:"w"`
	H          float64 `json:"h"`
	Lat        float64 `json:"lat"`
	Lon        float64 `json:"lon"`
	Alt        float64 `json:"alt"`
	Confidence float64 `json:"confidence"`
	TrackID    int     `json:"track_id"`
	Timestamp  float64 `json:"timestamp"`
}

// Video frame cache for capturing
type VideoFrameCache struct {
	frame     []byte
	timestamp float64
	mutex     sync.RWMutex
}

// Video stream hub for broadcasting frames to all clients
type VideoHub struct {
	clients    map[*VideoClient]bool
	broadcast  chan *VideoFrameMessage
	register   chan *VideoClient
	unregister chan *VideoClient
	mutex      sync.RWMutex
}

// Video client structure
type VideoClient struct {
	conn *websocket.Conn
	send chan []byte
}

// Attack hub for broadcasting attack data to all clients
type AttackHub struct {
	clients    map[*AttackClient]bool
	broadcast  chan *models.Attack
	register   chan *AttackClient
	unregister chan *AttackClient
	mutex      sync.RWMutex
}

// Attack client structure
type AttackClient struct {
	conn *websocket.Conn
	send chan []byte
}

// Global video hub instance
var videoHub *VideoHub

// Global attack hub instance
var attackHub *AttackHub

// Global hub instance
var hub *Hub

// Global video frame cache for MQTT detection captures
var videoFrameCache *VideoFrameCache

func init() {
	hub = &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *BroadcastMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
	go hub.run()

	// Initialize video hub
	videoHub = &VideoHub{
		clients:    make(map[*VideoClient]bool),
		broadcast:  make(chan *VideoFrameMessage, 100),
		register:   make(chan *VideoClient),
		unregister: make(chan *VideoClient),
	}
	go videoHub.run()

	// Initialize attack hub
	attackHub = &AttackHub{
		clients:    make(map[*AttackClient]bool),
		broadcast:  make(chan *models.Attack, 100),
		register:   make(chan *AttackClient),
		unregister: make(chan *AttackClient),
	}
	go attackHub.run()

	// Initialize video frame cache
	videoFrameCache = &VideoFrameCache{}
}

// Run video hub to handle video client connections and broadcasts
func (vh *VideoHub) run() {
	for {
		select {
		case client := <-vh.register:
			vh.mutex.Lock()
			vh.clients[client] = true
			vh.mutex.Unlock()
			log.Printf("Video client connected. Total clients: %d", len(vh.clients))

		case client := <-vh.unregister:
			vh.mutex.Lock()
			if _, ok := vh.clients[client]; ok {
				delete(vh.clients, client)
				close(client.send)
				log.Printf("Video client disconnected. Total clients: %d", len(vh.clients))
			}
			vh.mutex.Unlock()

		case frame := <-vh.broadcast:
			vh.mutex.RLock()
			frameData := mustMarshal(frame)
			// Collect clients to unregister
			var toUnregister []*VideoClient
			for client := range vh.clients {
				select {
				case client.send <- frameData:
				default:
					// Client buffer full, mark for disconnect
					toUnregister = append(toUnregister, client)
				}
			}
			vh.mutex.RUnlock()

			// Unregister slow clients outside the lock
			for _, client := range toUnregister {
				vh.unregister <- client
			}
		}
	}
}

// Run attack hub to handle attack client connections and broadcasts
func (ah *AttackHub) run() {
	for {
		select {
		case client := <-ah.register:
			ah.mutex.Lock()
			ah.clients[client] = true
			ah.mutex.Unlock()
			log.Printf("Attack client connected. Total clients: %d", len(ah.clients))

		case client := <-ah.unregister:
			ah.mutex.Lock()
			if _, ok := ah.clients[client]; ok {
				delete(ah.clients, client)
				close(client.send)
				log.Printf("Attack client disconnected. Total clients: %d", len(ah.clients))
			}
			ah.mutex.Unlock()

		case attack := <-ah.broadcast:
			ah.mutex.RLock()
			attackData := mustMarshal(attack)
			// Collect clients to unregister
			var toUnregister []*AttackClient
			for client := range ah.clients {
				select {
				case client.send <- attackData:
				default:
					// Client buffer full, mark for disconnect
					toUnregister = append(toUnregister, client)
				}
			}
			ah.mutex.RUnlock()

			// Unregister slow clients outside the lock
			for _, client := range toUnregister {
				ah.unregister <- client
			}
			log.Printf("Broadcasted attack data to %d clients. DroneID: %s", len(ah.clients), attack.DroneID)
		}
	}
}

// Broadcast video frame to all connected clients
func BroadcastVideoFrame(frame *VideoFrameMessage) {
	if videoHub != nil {
		select {
		case videoHub.broadcast <- frame:
			// Update frame cache for MQTT captures
			UpdateVideoFrameCache(frame)
		default:
			log.Println("Video broadcast channel full, dropping frame")
		}
	}
}

// UpdateVideoFrameCache updates the cached video frame
func UpdateVideoFrameCache(frame *VideoFrameMessage) {
	if videoFrameCache == nil {
		return
	}

	videoFrameCache.mutex.Lock()
	defer videoFrameCache.mutex.Unlock()

	// Decode base64 frame
	frameData, err := base64.StdEncoding.DecodeString(frame.Frame)
	if err != nil {
		log.Printf("Failed to decode video frame: %v", err)
		return
	}

	videoFrameCache.frame = frameData
	videoFrameCache.timestamp = frame.Timestamp
}

// GetLatestVideoFrame returns the latest cached video frame
func GetLatestVideoFrame() ([]byte, float64, error) {
	if videoFrameCache == nil {
		return nil, 0, fmt.Errorf("video frame cache not initialized")
	}

	videoFrameCache.mutex.RLock()
	defer videoFrameCache.mutex.RUnlock()

	if videoFrameCache.frame == nil {
		return nil, 0, fmt.Errorf("no video frame available")
	}

	// Return a copy of the frame
	frameCopy := make([]byte, len(videoFrameCache.frame))
	copy(frameCopy, videoFrameCache.frame)

	return frameCopy, videoFrameCache.timestamp, nil
}

// BroadcastAttack broadcasts attack data to all connected WebSocket clients
func BroadcastAttack(attack *models.Attack) {
	if attackHub != nil {
		select {
		case attackHub.broadcast <- attack:
		default:
			log.Println("Attack broadcast channel full, dropping attack data")
		}
	}
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
			// Collect clients to unregister
			var toUnregister []*Client
			for client := range h.clients {
				// Only send to clients subscribed to this camera
				if client.cameraID == message.CameraID {
					// Create message with image data
					detectionMsg := createDetectionMessage(message.Detect)
					select {
					case client.send <- mustMarshal(detectionMsg):
					default:
						// Client buffer full, mark for disconnect
						toUnregister = append(toUnregister, client)
					}
				}
			}
			h.mutex.RUnlock()

			// Unregister slow clients outside the lock
			for _, client := range toUnregister {
				h.unregister <- client
			}
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

// HandleAttackWebSocket - WebSocket handler for broadcasting attack data to frontend
func (h *detectHandler) HandleAttackWebSocket() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		client := &AttackClient{
			conn: c,
			send: make(chan []byte, 256),
		}

		// Register client
		attackHub.register <- client

		// Create channels for write operations
		writeChan := make(chan interface{}, 100)
		done := make(chan struct{})
		var closeOnce sync.Once

		// Start single write goroutine
		go func() {
			defer func() {
				attackHub.unregister <- client
			}()

			// Send initial confirmation
			if err := c.WriteJSON(fiber.Map{
				"status":  "connected",
				"message": "Subscribed to attack data stream",
			}); err != nil {
				log.Printf("Error sending initial message: %v", err)
				closeOnce.Do(func() { close(done) })
				return
			}

			for {
				select {
				case message, ok := <-client.send:
					if !ok {
						closeOnce.Do(func() { close(done) })
						return
					}
					if err := c.WriteMessage(websocket.TextMessage, message); err != nil {
						log.Printf("Error writing attack data: %v", err)
						closeOnce.Do(func() { close(done) })
						return
					}
				case msg := <-writeChan:
					switch m := msg.(type) {
					case []byte:
						if err := c.WriteMessage(websocket.PongMessage, m); err != nil {
							log.Printf("Error writing pong: %v", err)
							closeOnce.Do(func() { close(done) })
							return
						}
					case map[string]interface{}:
						if err := c.WriteJSON(m); err != nil {
							log.Printf("Error writing JSON: %v", err)
							closeOnce.Do(func() { close(done) })
							return
						}
					}
				case <-done:
					return
				}
			}
		}()

		// Keep connection alive and handle ping/pong
		for {
			messageType, payload, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("Attack client unexpected close: %v", err)
				}
				closeOnce.Do(func() { close(done) })
				break
			}

			// Handle ping - send pong via write channel
			if messageType == websocket.PingMessage {
				select {
				case writeChan <- []byte(nil):
				default:
					log.Printf("Write channel full, skipping pong")
				}
			}

			// Handle text messages (e.g., ping from client)
			if messageType == websocket.TextMessage {
				var msg map[string]interface{}
				if err := json.Unmarshal(payload, &msg); err == nil {
					if msgType, ok := msg["type"].(string); ok && msgType == "ping" {
						// Send pong response via write channel
						select {
						case writeChan <- map[string]interface{}{
							"type":      "pong",
							"timestamp": msg["timestamp"],
						}:
						default:
							log.Printf("Write channel full, skipping pong")
						}
					}
				}
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

// HandleVideoInput - WebSocket handler for receiving video frames from Python
func (h *detectHandler) HandleVideoInput() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		log.Println("Python video source connected")
		defer func() {
			log.Println("Python video source disconnected")
			c.Close()
		}()

		// Send confirmation
		c.WriteJSON(fiber.Map{
			"status":  "connected",
			"message": "Ready to receive video frames",
		})

		// Read frames from Python
		for {
			var frame VideoFrameMessage
			if err := c.ReadJSON(&frame); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("Unexpected close error: %v", err)
				}
				break
			}

			// Broadcast frame to all viewer clients
			BroadcastVideoFrame(&frame)

			// Log every 30 frames
			if frame.FrameNumber%30 == 0 {
				log.Printf("Received frame #%d, detections: %d, viewers: %d",
					frame.FrameNumber, frame.Detections, len(videoHub.clients))
			}
		}
	})
}

// HandleVideoStream - WebSocket handler for clients to view the video stream
func (h *detectHandler) HandleVideoStream() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		client := &VideoClient{
			conn: c,
			send: make(chan []byte, 256),
		}

		// Register client
		videoHub.register <- client

		// Create channels for write operations
		writeChan := make(chan interface{}, 100)
		done := make(chan struct{})
		var closeOnce sync.Once

		// Start single write goroutine
		go func() {
			defer func() {
				videoHub.unregister <- client
			}()

			// Send initial confirmation
			if err := c.WriteJSON(fiber.Map{
				"status":  "connected",
				"message": "Subscribed to video stream",
			}); err != nil {
				log.Printf("Error sending initial message: %v", err)
				closeOnce.Do(func() { close(done) })
				return
			}

			for {
				select {
				case message, ok := <-client.send:
					if !ok {
						closeOnce.Do(func() { close(done) })
						return
					}
					if err := c.WriteMessage(websocket.TextMessage, message); err != nil {
						log.Printf("Error writing video frame: %v", err)
						closeOnce.Do(func() { close(done) })
						return
					}
				case msg := <-writeChan:
					switch m := msg.(type) {
					case []byte:
						if err := c.WriteMessage(websocket.PongMessage, m); err != nil {
							log.Printf("Error writing pong: %v", err)
							closeOnce.Do(func() { close(done) })
							return
						}
					case map[string]interface{}:
						if err := c.WriteJSON(m); err != nil {
							log.Printf("Error writing JSON: %v", err)
							closeOnce.Do(func() { close(done) })
							return
						}
					}
				case <-done:
					return
				}
			}
		}()

		// Keep connection alive and handle ping/pong
		for {
			messageType, payload, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("Video client unexpected close: %v", err)
				}
				closeOnce.Do(func() { close(done) })
				break
			}

			// Handle ping - send pong via write channel
			if messageType == websocket.PingMessage {
				select {
				case writeChan <- []byte(nil):
				default:
					log.Printf("Write channel full, skipping pong")
				}
			}

			// Handle text messages (e.g., ping from client)
			if messageType == websocket.TextMessage {
				var msg map[string]interface{}
				if err := json.Unmarshal(payload, &msg); err == nil {
					if msgType, ok := msg["type"].(string); ok && msgType == "ping" {
						// Send pong response via write channel
						select {
						case writeChan <- map[string]interface{}{
							"type":      "pong",
							"timestamp": msg["timestamp"],
						}:
						default:
							log.Printf("Write channel full, skipping pong")
						}
					}
				}
			}
		}
	})
}
