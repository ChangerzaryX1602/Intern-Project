package detect

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"topgun-services/pkg/domain"
	"topgun-services/pkg/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	helpers "github.com/zercle/gofiber-helpers"
)

type detectHandler struct {
	service domain.DetectService
}

func NewDetectHandler(router fiber.Router, service domain.DetectService) {
	h := &detectHandler{service: service}

	// WebSocket route
	router.Get("/ws", WebSocketUpgrade(), h.HandleWebSocket())

	// HTTP routes
	router.Post("/", h.CreateDetect())
	router.Get("/", h.GetDetects())
	router.Get("/:id", h.GetDetect())
	router.Get("/:id/file", h.GetDetectFile())
	router.Put("/:id", h.UpdateDetect())
	router.Delete("/:id", h.DeleteDetect())
}

// @Summary CreateDetect
// @Tags Detect
// @Description Create a new detect with file upload
// @Accept multipart/form-data
// @Produce json
// @Param camera_id formData string true "Camera ID (UUID)"
// @Param objects formData string false "Objects JSON array"
// @Param file formData file true "File to upload"
// @Router /api/v1/detect/ [post]
// @Security ApiKeyAuth
func (h *detectHandler) CreateDetect() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse camera_id from form
		cameraIDStr := c.FormValue("camera_id")
		if cameraIDStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Title:   "Missing camera_id",
						Message: "camera_id is required",
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}

		cameraID, err := uuid.Parse(cameraIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Title:   "Invalid camera_id",
						Message: "camera_id must be a valid UUID",
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}

		// Get uploaded file
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Title:   "Missing file",
						Message: "file is required",
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}

		// Create upload directory if not exists
		uploadDir := "upload"
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusInternalServerError,
						Title:   "Failed to create upload directory",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}

		// Generate UUID for filename
		fileExt := filepath.Ext(file.Filename)
		newFilename := uuid.New().String() + fileExt
		filePath := filepath.Join(uploadDir, newFilename)

		// Save file
		if err := c.SaveFile(file, filePath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusInternalServerError,
						Title:   "Failed to save file",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}

		// Parse objects if provided
		var objects models.JSONRawMessageArray
		objectsStr := c.FormValue("objects")
		if objectsStr != "" {
			if err := json.Unmarshal([]byte(objectsStr), &objects); err != nil {
				// Clean up uploaded file if JSON parsing fails
				os.Remove(filePath)
				return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
					Success: false,
					Errors: []helpers.ResponseError{
						{
							Code:    fiber.StatusBadRequest,
							Title:   "Invalid objects JSON",
							Message: err.Error(),
							Source:  helpers.WhereAmI(),
						},
					},
				})
			}
		}

		// Create detect object
		detect := models.Detect{
			CameraID: cameraID,
			Path:     filePath,
			Objects:  objects,
		}

		createdDetect, err := h.service.CreateDetect(detect)
		if err != nil {
			// Clean up uploaded file if database operation fails
			os.Remove(filePath)
			return c.Status(fiber.StatusInternalServerError).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusInternalServerError,
						Title:   "Failed to create detect",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}

		// Broadcast detection to WebSocket clients
		BroadcastDetection(createdDetect)

		return c.Status(fiber.StatusCreated).JSON(helpers.ResponseForm{
			Success: true,
			Data:    createdDetect,
		})
	}
}

// @Summary GetDetects
// @Tags Detect
// @Description Get a list of detects with pagination and filtering
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param search query string false "Search filter"
// @Router /api/v1/detect/ [get]
// @Security ApiKeyAuth
func (h *detectHandler) GetDetects() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var pagination models.Pagination
		var filter models.Search

		if err := c.QueryParser(&pagination); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Title:   "Invalid pagination query parameters",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}
		if err := c.QueryParser(&filter); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Title:   "Invalid filter query parameters",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}

		detects, p, s, err := h.service.GetDetects(pagination, filter)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusInternalServerError,
						Title:   "Failed to retrieve detects",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}

		return c.Status(fiber.StatusOK).JSON(helpers.ResponseForm{
			Success: true,
			Data: fiber.Map{
				"detects":    detects,
				"pagination": p,
				"search":     s,
			},
		})
	}
}

// @Summary GetDetect
// @Tags Detect
// @Description Get a detect by ID
// @Accept json
// @Produce json
// @Param id path string true "Detect ID"
// @Router /api/v1/detect/{id} [get]
// @Security ApiKeyAuth
func (h *detectHandler) GetDetect() fiber.Handler {
	return func(c *fiber.Ctx) error {
		idParam := c.Params("id")
		var id uint
		_, err := fmt.Sscan(idParam, &id)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Title:   "Invalid detect ID",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}

		detect, err := h.service.GetDetect(id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusInternalServerError,
						Title:   "Failed to retrieve detect",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}

		return c.Status(fiber.StatusOK).JSON(helpers.ResponseForm{
			Success: true,
			Data:    detect,
		})
	}
}

// @Summary GetDetectFile
// @Tags Detect
// @Description Download the file associated with a detect by ID
// @Produce application/octet-stream
// @Param id path string true "Detect ID"
// @Router /api/v1/detect/{id}/file [get]
// @Security ApiKeyAuth
func (h *detectHandler) GetDetectFile() fiber.Handler {
	return func(c *fiber.Ctx) error {
		idParam := c.Params("id")
		var id uint
		_, err := fmt.Sscan(idParam, &id)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Title:   "Invalid detect ID",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}

		detect, err := h.service.GetDetectFile(id)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusNotFound,
						Title:   "Detect not found",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}

		// Check if file exists
		if detect.Path == "" {
			return c.Status(fiber.StatusNotFound).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusNotFound,
						Title:   "File not found",
						Message: "No file associated with this detect",
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}

		// Check if file exists on disk
		if _, err := os.Stat(detect.Path); os.IsNotExist(err) {
			return c.Status(fiber.StatusNotFound).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusNotFound,
						Title:   "File not found",
						Message: "File does not exist on server",
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}

		// Get filename from path
		filename := filepath.Base(detect.Path)

		// Set headers for file download
		c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.Set("Content-Type", "application/octet-stream")

		// Send file
		return c.SendFile(detect.Path)
	}
}

// @Summary UpdateDetect
// @Tags Detect
// @Description Update an existing detect
// @Accept json
// @Produce json
// @Param id path string true "Detect ID"
// @Param detect body models.Detect true "Detect object"
// @Router /api/v1/detect/{id} [put]
// @Security ApiKeyAuth
func (h *detectHandler) UpdateDetect() fiber.Handler {
	return func(c *fiber.Ctx) error {
		idParam := c.Params("id")
		var id uint
		_, err := fmt.Sscan(idParam, &id)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Title:   "Invalid detect ID",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}

		var detect models.Detect
		if err := c.BodyParser(&detect); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Title:   "Invalid request body",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}

		updatedDetect, err := h.service.UpdateDetect(id, detect)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusInternalServerError,
						Title:   "Failed to update detect",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}

		return c.Status(fiber.StatusOK).JSON(helpers.ResponseForm{
			Success: true,
			Data:    updatedDetect,
		})
	}
}

// @Summary DeleteDetect
// @Tags Detect
// @Description Delete a detect by ID
// @Accept json
// @Produce json
// @Param id path string true "Detect ID"
// @Router /api/v1/detect/{id} [delete]
// @Security ApiKeyAuth
func (h *detectHandler) DeleteDetect() fiber.Handler {
	return func(c *fiber.Ctx) error {
		idParam := c.Params("id")
		var id uint
		_, err := fmt.Sscan(idParam, &id)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Title:   "Invalid detect ID",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}

		err = h.service.DeleteDetect(id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusInternalServerError,
						Title:   "Failed to delete detect",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}

		return c.Status(fiber.StatusOK).JSON(helpers.ResponseForm{
			Success: true,
			Data:    nil,
		})
	}
}
