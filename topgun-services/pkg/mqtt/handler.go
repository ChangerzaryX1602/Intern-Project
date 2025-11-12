package mqtt

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// PublishMessage handles POST requests to publish messages to MQTT
// @Summary Publish message to MQTT
// @Description Publish a message to the configured MQTT topic (topgun/ai)
// @Tags MQTT
// @Accept json
// @Produce json
// @Param message body PublishRequest true "Message to publish"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/mqtt/publish [post]
func (h *Handler) PublishMessage(c *fiber.Ctx) error {
	var req PublishRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Message cannot be empty",
		})
	}

	if err := h.service.Publish(req.Message); err != nil {
		log.Printf("Error publishing message: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to publish message",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Message published successfully",
		"topic":   h.service.GetTopic(),
	})
}

// PublishJSON handles POST requests to publish JSON data to MQTT
// @Summary Publish JSON to MQTT
// @Description Publish JSON data to the configured MQTT topic (topgun/ai)
// @Tags MQTT
// @Accept json
// @Produce json
// @Param data body map[string]interface{} true "JSON data to publish"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/mqtt/publish-json [post]
func (h *Handler) PublishJSON(c *fiber.Ctx) error {
	var data map[string]interface{}
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON body",
		})
	}

	if err := h.service.PublishJSON(data); err != nil {
		log.Printf("Error publishing JSON: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to publish JSON",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "JSON published successfully",
		"topic":   h.service.GetTopic(),
	})
}

// GetStatus handles GET requests to check MQTT connection status
// @Summary Get MQTT connection status
// @Description Check if the MQTT client is connected to the broker
// @Tags MQTT
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/mqtt/status [get]
func (h *Handler) GetStatus(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"connected": h.service.IsConnected(),
		"topic":     h.service.GetTopic(),
	})
}

type PublishRequest struct {
	Message string `json:"message" example:"Hello from Go server"`
}

// UploadFile handles file upload and publishes it via MQTT
// @Summary Upload and send file via MQTT
// @Description Upload a file (e.g., .pt model file) and send it through MQTT topic
// @Tags MQTT
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload and send (.pt, .pth, etc.)"
// @Param encode_base64 formData boolean false "Encode file as base64 (default: true)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/mqtt/upload-file [post]
func (h *Handler) UploadFile(c *fiber.Ctx) error {
	// Get the file from form data
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No file provided",
		})
	}

	// Check if we should encode as base64 (default: true)
	encodeBase64 := c.FormValue("encode_base64", "true") == "true"

	// Open the file
	fileContent, err := file.Open()
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to open file",
		})
	}
	defer fileContent.Close()

	// Read file content
	fileBytes, err := io.ReadAll(fileContent)
	if err != nil {
		log.Printf("Error reading file: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read file",
		})
	}

	// Get file extension
	ext := filepath.Ext(file.Filename)

	// Prepare metadata
	metadata := map[string]interface{}{
		"type":     "file_upload",
		"filename": file.Filename,
		"size":     file.Size,
		"ext":      ext,
		"encoded":  encodeBase64,
	}

	var dataToSend interface{}

	if encodeBase64 {
		// Encode file as base64 for text-safe transmission
		encoded := base64.StdEncoding.EncodeToString(fileBytes)
		dataToSend = map[string]interface{}{
			"metadata": metadata,
			"data":     encoded,
		}

		// Publish as JSON
		if err := h.service.PublishJSON(dataToSend); err != nil {
			log.Printf("Error publishing file: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to publish file",
				"details": err.Error(),
			})
		}
	} else {
		// First send metadata
		if err := h.service.PublishFileMetadata(metadata); err != nil {
			log.Printf("Error publishing metadata: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to publish metadata",
				"details": err.Error(),
			})
		}

		// Then send raw bytes
		if err := h.service.PublishBytes(fileBytes); err != nil {
			log.Printf("Error publishing file bytes: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to publish file bytes",
				"details": err.Error(),
			})
		}
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"message":  "File uploaded and published successfully",
		"topic":    h.service.GetTopic(),
		"filename": file.Filename,
		"size":     file.Size,
		"encoded":  encodeBase64,
	})
}

// SendFileByPath handles sending a file from server filesystem via MQTT
// @Summary Send file from server path via MQTT
// @Description Send a file from the server filesystem through MQTT topic
// @Tags MQTT
// @Accept json
// @Produce json
// @Param request body FilePathRequest true "File path on server"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/mqtt/send-file [post]
func (h *Handler) SendFileByPath(c *fiber.Ctx) error {
	var req FilePathRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.FilePath == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File path is required",
		})
	}

	// Read file from server
	fileBytes, err := os.ReadFile(req.FilePath)
	if err != nil {
		log.Printf("Error reading file from path: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   fmt.Sprintf("Failed to read file: %s", req.FilePath),
			"details": err.Error(),
		})
	}

	filename := filepath.Base(req.FilePath)
	ext := filepath.Ext(filename)
	encodeBase64 := req.EncodeBase64

	// Prepare metadata
	metadata := map[string]interface{}{
		"type":     "file_send",
		"filename": filename,
		"size":     len(fileBytes),
		"ext":      ext,
		"path":     req.FilePath,
		"encoded":  encodeBase64,
	}

	if encodeBase64 {
		// Encode and send as JSON
		encoded := base64.StdEncoding.EncodeToString(fileBytes)
		dataToSend := map[string]interface{}{
			"metadata": metadata,
			"data":     encoded,
		}

		if err := h.service.PublishJSON(dataToSend); err != nil {
			log.Printf("Error publishing file: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to publish file",
				"details": err.Error(),
			})
		}
	} else {
		// Send metadata first, then raw bytes
		if err := h.service.PublishFileMetadata(metadata); err != nil {
			log.Printf("Error publishing metadata: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to publish metadata",
				"details": err.Error(),
			})
		}

		if err := h.service.PublishBytes(fileBytes); err != nil {
			log.Printf("Error publishing file bytes: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to publish file bytes",
				"details": err.Error(),
			})
		}
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"message":  "File sent successfully",
		"topic":    h.service.GetTopic(),
		"filename": filename,
		"size":     len(fileBytes),
		"encoded":  encodeBase64,
	})
}

type FilePathRequest struct {
	FilePath     string `json:"file_path" example:"./models/best.pt"`
	EncodeBase64 bool   `json:"encode_base64" example:"true"`
}

// SetupRoutes sets up the MQTT routes
func SetupRoutes(group fiber.Router, handler *Handler) {
	group.Get("/status", handler.GetStatus)
	group.Post("/publish", handler.PublishMessage)
	group.Post("/publish-json", handler.PublishJSON)
	group.Post("/upload-file", handler.UploadFile)
	group.Post("/send-file", handler.SendFileByPath)
}
