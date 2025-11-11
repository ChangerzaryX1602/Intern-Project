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
	router.Get("/attacks", handler.GetAttacks())
	router.Post("/attacks", handler.CreateAttack())
	router.Put("/attacks/:id", handler.UpdateAttack())
	router.Delete("/attacks/:id", handler.DeleteAttack())
	router.Get("/attacks/:id", handler.GetAttack())
}
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
