package attack

import (
	"topgun-services/pkg/domain"
	"topgun-services/pkg/models"

	"github.com/gofiber/fiber/v2"
	helpers "github.com/zercle/gofiber-helpers"
)

type attackHandler struct {
	service domain.AttackService
}

func NewAttackHandler(router fiber.Router, service domain.AttackService) {
	handler := &attackHandler{service: service}
	router.Get("/", handler.GetAttacks())
	router.Post("/", handler.CreateAttack())
	router.Put("/:id", handler.UpdateAttack())
	router.Delete("/:id", handler.DeleteAttack())
	router.Get("/:id", handler.GetAttack())
}

// @Summary Get Attacks
// @Description Retrieve a list of attacks with pagination and filtering options.
// @Tags Attacks
// @Accept json
// @Produce json
// @Param page query int false "Page number for pagination"
// @Param per_page query int false "Number of items per page"
// @Param keyword query string false "Keyword to filter attacks"
// @Param column query string false "Column to sort by"
// @Router /api/v1/attack [get]
func (h *attackHandler) GetAttacks() fiber.Handler {
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
		attacks, pag, fil, err := h.service.GetAttacks(pagination, filter)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusInternalServerError,
						Title:   "Failed to get attacks",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}
		return c.Status(fiber.StatusOK).JSON(helpers.ResponseForm{
			Success: true,
			Data: fiber.Map{
				"attacks":    attacks,
				"pagination": pag,
				"filter":     fil,
			},
		})
	}
}

// @Summary Create Attack
// @Description Create a new attack record.
// @Tags Attacks
// @Accept json
// @Produce json
// @Param attack body models.Attack true "Attack data"
// @Router /api/v1/attack [post]
func (h *attackHandler) CreateAttack() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var attack models.Attack
		if err := c.BodyParser(&attack); err != nil {
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
		createdAttack, err := h.service.CreateAttack(attack)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusInternalServerError,
						Title:   "Failed to create attack",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}
		return c.Status(fiber.StatusCreated).JSON(helpers.ResponseForm{
			Success: true,
			Data:    createdAttack,
		})
	}
}

// @Summary Update Attack
// @Description Update an existing attack record by ID.
// @Tags Attacks
// @Accept json
// @Produce json
// @Param id path int true "Attack ID"
// @Param attack body models.Attack true "Updated attack data"
// @Router /api/v1/attack/{id} [put]
func (h *attackHandler) UpdateAttack() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Title:   "Invalid attack ID parameter",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}
		var attack models.Attack
		if err := c.BodyParser(&attack); err != nil {
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
		attack.ID = uint(id)
		updatedAttack, err := h.service.UpdateAttack(uint(id), attack)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusInternalServerError,
						Title:   "Failed to update attack",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}
		return c.Status(fiber.StatusOK).JSON(helpers.ResponseForm{
			Success: true,
			Data:    updatedAttack,
		})
	}
}

// @Summary Delete Attack
// @Description Delete an existing attack record by ID.
// @Tags Attacks
// @Accept json
// @Produce json
// @Param id path int true "Attack ID"
// @Router /api/v1/attack/{id} [delete]
func (h *attackHandler) DeleteAttack() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Title:   "Invalid attack ID parameter",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}
		err = h.service.DeleteAttack(uint(id))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusInternalServerError,
						Title:   "Failed to delete attack",
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

// @Summary Get Attack
// @Description Retrieve a single attack record by ID.
// @Tags Attacks
// @Accept json
// @Produce json
// @Param id path int true "Attack ID"
// @Router /api/v1/attack/{id} [get]
func (h *attackHandler) GetAttack() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Title:   "Invalid attack ID parameter",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}
		attack, err := h.service.GetAttack(uint(id))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusInternalServerError,
						Title:   "Failed to get attack",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}
		return c.Status(fiber.StatusOK).JSON(helpers.ResponseForm{
			Success: true,
			Data:    attack,
		})
	}
}
