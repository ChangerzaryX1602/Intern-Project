package detect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"time"
	"topgun-services/pkg/domain"
	"topgun-services/pkg/models"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/nfnt/resize"
)

// MQTTDetectHandler handles MQTT messages for detection data
type MQTTDetectHandler struct {
	service  domain.DetectService
	cameraID uuid.UUID
}

// NewMQTTDetectHandler creates a new MQTT detect handler
func NewMQTTDetectHandler(service domain.DetectService, cameraID uuid.UUID) *MQTTDetectHandler {
	return &MQTTDetectHandler{
		service:  service,
		cameraID: cameraID,
	}
}

// HandleMessage processes incoming MQTT messages
func (h *MQTTDetectHandler) HandleMessage(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received MQTT message on topic %s", msg.Topic())

	// Parse RaspberryPI detection data
	var piDetection RaspberryPIDetection
	if err := json.Unmarshal(msg.Payload(), &piDetection); err != nil {
		log.Printf("Failed to parse MQTT message: %v", err)
		return
	}

	log.Printf("RaspberryPI Detection: TrackID=%d, Lat=%.6f, Lon=%.6f, Alt=%.2f, Confidence=%.2f",
		piDetection.TrackID, piDetection.Lat, piDetection.Lon, piDetection.Alt, piDetection.Confidence)

	// Capture current video frame
	frameData, _, err := GetLatestVideoFrame()
	if err != nil {
		log.Printf("Failed to get video frame: %v", err)
		// Continue anyway, we'll save detection without image
	}

	// Save captured frame to file
	var imagePath string
	if frameData != nil {
		imagePath, err = h.saveFrameToFile(frameData, piDetection.TrackID)
		if err != nil {
			log.Printf("Failed to save frame to file: %v", err)
		} else {
			log.Printf("Saved captured frame to: %s", imagePath)
		}
	}

	// Create objects array with detection data
	objectsData := map[string]interface{}{
		"x":          piDetection.X,
		"y":          piDetection.Y,
		"w":          piDetection.W,
		"h":          piDetection.H,
		"lat":        piDetection.Lat,
		"lon":        piDetection.Lon,
		"alt":        piDetection.Alt,
		"confidence": piDetection.Confidence,
		"track_id":   piDetection.TrackID,
		"timestamp":  piDetection.Timestamp,
	}

	// Marshal to JSON
	objectsJSON, err := json.Marshal([]interface{}{objectsData})
	if err != nil {
		log.Printf("Failed to marshal objects data: %v", err)
		return
	}

	// Create detection record
	detect := models.Detect{
		CameraID:  h.cameraID,
		Timestamp: time.Unix(int64(piDetection.Timestamp), 0),
		Path:      imagePath,
	}

	// Parse objects into JSONRawMessageArray
	if err := json.Unmarshal(objectsJSON, &detect.Objects); err != nil {
		log.Printf("Failed to unmarshal objects: %v", err)
		return
	}

	// Save to database
	savedDetect, err := h.service.CreateDetect(detect)
	if err != nil {
		log.Printf("Failed to save detection to database: %v", err)
		return
	}

	log.Printf("Successfully saved detection ID=%d with %d objects to database", savedDetect.ID, len(savedDetect.Objects))

	// Broadcast to WebSocket clients
	if hub != nil {
		hub.broadcast <- &BroadcastMessage{
			CameraID: h.cameraID,
			Detect:   savedDetect,
		}
		log.Printf("Broadcasted detection to WebSocket clients")
	}
}

// saveFrameToFile saves the captured frame to upload directory
func (h *MQTTDetectHandler) saveFrameToFile(frameData []byte, trackID int) (string, error) {
	// Create upload directory if not exists
	uploadDir := "./upload"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Decode image from JPEG bytes
	img, err := jpeg.Decode(bytes.NewReader(frameData))
	if err != nil {
		return "", fmt.Errorf("failed to decode JPEG image: %w", err)
	}

	// Resize to 720p (1280x720) while maintaining aspect ratio
	// If original is smaller than 720p, keep original size
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var resizedImg image.Image
	if width > 1280 || height > 720 {
		// Calculate target dimensions maintaining aspect ratio
		aspectRatio := float64(width) / float64(height)
		var targetWidth, targetHeight uint

		if aspectRatio > 16.0/9.0 {
			// Width is the limiting factor
			targetWidth = 1280
			targetHeight = 0 // Let resize calculate height
		} else {
			// Height is the limiting factor
			targetWidth = 0 // Let resize calculate width
			targetHeight = 720
		}

		// Resize using Lanczos3 (high quality)
		resizedImg = resize.Resize(targetWidth, targetHeight, img, resize.Lanczos3)
		log.Printf("Resized image from %dx%d to %dx%d", width, height, resizedImg.Bounds().Dx(), resizedImg.Bounds().Dy())
	} else {
		resizedImg = img
		log.Printf("Image size %dx%d is already smaller than 720p, keeping original", width, height)
	}

	// Generate filename with timestamp and track_id
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("mqtt_capture_%s_track_%d.jpg", timestamp, trackID)
	filePath := filepath.Join(uploadDir, filename)

	// Create output file
	outFile, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Encode and save as JPEG with quality 90
	if err := jpeg.Encode(outFile, resizedImg, &jpeg.Options{Quality: 90}); err != nil {
		return "", fmt.Errorf("failed to encode JPEG: %w", err)
	}

	return filePath, nil
}

// StartMQTTSubscription starts subscribing to MQTT topic for detection data
func StartMQTTSubscription(mqttBroker, mqttTopic string, cameraID uuid.UUID, service domain.DetectService) error {
	// Create MQTT client options
	opts := mqtt.NewClientOptions()
	opts.AddBroker(mqttBroker)
	opts.SetClientID(fmt.Sprintf("topgun-detect-%s", uuid.New().String()))
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(5 * time.Second)
	opts.SetMaxReconnectInterval(30 * time.Second)

	// Connection handlers
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		log.Printf("MQTT connected to broker: %s", mqttBroker)
	})

	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		log.Printf("MQTT connection lost: %v", err)
	})

	// Create MQTT client
	client := mqtt.NewClient(opts)

	// Connect to broker
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}

	// Create message handler
	handler := NewMQTTDetectHandler(service, cameraID)

	// Subscribe to topic
	if token := client.Subscribe(mqttTopic, 1, handler.HandleMessage); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to subscribe to MQTT topic %s: %w", mqttTopic, token.Error())
	}

	log.Printf("Successfully subscribed to MQTT topic: %s", mqttTopic)

	return nil
}
