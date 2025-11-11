package camera

import (
	"topgun-services/internal/handlers"
	"topgun-services/pkg/domain"
	"topgun-services/pkg/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	helpers "github.com/zercle/gofiber-helpers"
)

type cameraHandler struct {
	service domain.CameraService
}

func NewCameraHandler(router fiber.Router, routerResource *handlers.RouterResources, cameraService domain.CameraService) {
	handler := &cameraHandler{
		service: cameraService,
	}
	router.Get("/", handler.GetCameras())
	router.Get("/:id", handler.GetCamera())
	router.Post("/", handler.CreateCamera())
	router.Put("/:id", handler.UpdateCamera())
	router.Delete("/:id", handler.DeleteCamera())
}

// @Summary GetCameras
// @Tags Camera
// @Description Get list of cameras with pagination and search
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param search query string false "Search term"
// @Router /api/v1/camera/ [get]
// @Security ApiKeyAuth
func (h *cameraHandler) GetCameras() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var pagination models.Pagination
		var search models.Search
		if err := c.QueryParser(&pagination); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Title:   "Invalid pagination parameters",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}
		if err := c.QueryParser(&search); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Title:   "Invalid search parameters",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}
		cameras, p, s, err := h.service.GetCameras(pagination, search)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusInternalServerError,
						Title:   "Failed to retrieve cameras",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}
		return c.Status(fiber.StatusOK).JSON(helpers.ResponseForm{
			Success: true,
			Data: fiber.Map{
				"cameras":    cameras,
				"pagination": p,
				"search":     s,
			},
		})
	}
}

// @Summary GetCamera
// @Tags Camera
// @Description Get a camera by ID
// @Accept json
// @Produce json
// @Param id path string true "Camera ID"
// @Router /api/v1/camera/{id} [get]
// @Security ApiKeyAuth
func (h *cameraHandler) GetCamera() fiber.Handler {
	return func(c *fiber.Ctx) error {
		idParam := c.Params("id")
		id, err := uuid.Parse(idParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Title:   "Invalid camera ID",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}
		camera, err := h.service.GetCamera(id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusInternalServerError,
						Title:   "Failed to retrieve camera",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}
		return c.Status(fiber.StatusOK).JSON(helpers.ResponseForm{
			Success: true,
			Data:    camera,
		})
	}
}

// @Summary CreateCamera
// @Tags Camera
// @Description Create a new camera
// @Accept json
// @Produce json
// @Param camera body models.Camera true "Camera object"
// @Router /api/v1/camera/ [post]
// @Security ApiKeyAuth
func (h *cameraHandler) CreateCamera() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var camera models.Camera
		if err := c.BodyParser(&camera); err != nil {
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
		createdCamera, err := h.service.CreateCamera(camera)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusInternalServerError,
						Title:   "Failed to create camera",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}
		return c.Status(fiber.StatusCreated).JSON(helpers.ResponseForm{
			Success: true,
			Data:    createdCamera,
		})
	}
}

// @Summary UpdateCamera
// @Tags Camera
// @Description Update an existing camera
// @Accept json
// @Produce json
// @Param id path string true "Camera ID"
// @Param camera body models.Camera true "Camera object"
// @Router /api/v1/camera/{id} [put]
// @Security ApiKeyAuth
func (h *cameraHandler) UpdateCamera() fiber.Handler {
	return func(c *fiber.Ctx) error {
		idParam := c.Params("id")
		id, err := uuid.Parse(idParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Title:   "Invalid camera ID",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}
		var camera models.Camera
		if err := c.BodyParser(&camera); err != nil {
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
		updatedCamera, err := h.service.UpdateCamera(id, camera)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusInternalServerError,
						Title:   "Failed to update camera",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}
		return c.Status(fiber.StatusOK).JSON(helpers.ResponseForm{
			Success: true,
			Data:    updatedCamera,
		})
	}
}

// @Summary DeleteCamera
// @Tags Camera
// @Description Delete a camera by ID
// @Accept json
// @Produce json
// @Param id path string true "Camera ID"
// @Router /api/v1/camera/{id} [delete]
// @Security ApiKeyAuth
func (h *cameraHandler) DeleteCamera() fiber.Handler {
	return func(c *fiber.Ctx) error {
		idParam := c.Params("id")
		id, err := uuid.Parse(idParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Title:   "Invalid camera ID",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}
		err = h.service.DeleteCamera(id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusInternalServerError,
						Title:   "Failed to delete camera",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}
		return c.Status(fiber.StatusNoContent).JSON(helpers.ResponseForm{
			Success: true,
			Data:    nil,
		})
	}
}
