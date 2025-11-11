package user

import (
	"topgun-services/internal/handlers"
	"topgun-services/pkg/domain"
	"topgun-services/pkg/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	helpers "github.com/zercle/gofiber-helpers"
)

type userHandler struct {
	userService domain.UserService
	authService domain.AuthService
}

func NewUserHandler(router fiber.Router, routerResource *handlers.RouterResources, userService domain.UserService, authService domain.AuthService) {
	handler := &userHandler{
		userService: userService,
		authService: authService,
	}
	router.Get("/", routerResource.ReqAuthHandler(), handler.GetUsers())
	router.Get("/me", routerResource.ReqAuthHandler(), handler.GetMe())
	router.Get("/:id", routerResource.ReqAuthHandler(), handler.GetUser())
	router.Post("/", routerResource.ReqAuthHandler(), handler.CreateUser())
	router.Put("/:id", routerResource.ReqAuthHandler(), handler.UpdateUser())
	router.Delete("/:id", routerResource.ReqAuthHandler(), handler.DeleteUser())
}

// @Summary GetMe
// @Tags User
// @Description Get me
// @Accept json
// @Produce json
// @Router /api/v1/users/me [get]
// @Security ApiKeyAuth
func (h *userHandler) GetMe() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Locals("user_id").(string)
		uuid, err := uuid.Parse(id)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Title:   "Cannot parse user id as uuid",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}

		user, err := h.userService.GetUser(uuid)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusInternalServerError,
						Title:   "Internal server error",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}

		return c.Status(fiber.StatusOK).JSON(helpers.ResponseForm{
			Success: true,
			Data:    user,
		})
	}
}

// @Summary GetUsers
// @Tags User
// @Description Get users
// @Accept json
// @Produce json
// @Param page query int false "Page" default(1)
// @Param per_page query int false "Per Page" default(10)
// @Param keyword query string false "Keyword"
// @Param column query string false "Column" default("first_name,last_name")
// @Router /api/v1/users/ [get]
// @Security ApiKeyAuth
func (h *userHandler) GetUsers() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var pagination models.Pagination
		var search models.Search
		if err := c.QueryParser(&pagination); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Title:   "Invalid pagination",
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
						Title:   "Invalid filter",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}

		users, p, s, err := h.userService.GetUsers(pagination, search)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusInternalServerError,
						Title:   "Internal server error",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}

		return c.Status(fiber.StatusOK).JSON(helpers.ResponseForm{
			Success: true,
			Data: fiber.Map{
				"users":      users,
				"pagination": p,
				"search":     s,
			},
		})
	}
}

// @Summary CreateUser
// @Tags User
// @Description Create user
// @Accept json
// @Produce json
// @Param user body models.User true "User"
// @Router /api/v1/users/ [post]
// @Security ApiKeyAuth
func (h *userHandler) CreateUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var user models.User
		if err := c.BodyParser(&user); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Title:   "Invalid user",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}

		u, err := h.userService.CreateUser(user)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusInternalServerError,
						Title:   "Internal server error",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}

		return c.Status(fiber.StatusCreated).JSON(helpers.ResponseForm{
			Success: true,
			Data:    u,
		})
	}
}

// @Summary UpdateUser
// @Tags User
// @Description Update user
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body models.User true "User"
// @Router /api/v1/users/{id} [put]
// @Security ApiKeyAuth
func (h *userHandler) UpdateUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		uuid, err := uuid.Parse(id)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Title:   "Invalid user id",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}
		var user models.User
		if err := c.BodyParser(&user); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Title:   "Invalid user",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}
		u, err := h.userService.UpdateUser(uuid, user)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusInternalServerError,
						Title:   "Internal server error",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}

		return c.Status(fiber.StatusOK).JSON(helpers.ResponseForm{
			Success: true,
			Data:    u,
		})
	}
}

// @Summary DeleteUser
// @Tags User
// @Description Delete user
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Router /api/v1/users/{id} [delete]
// @Security ApiKeyAuth
func (h *userHandler) DeleteUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		uuid, err := uuid.Parse(id)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Title:   "Invalid user id",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}

		if err := h.userService.DeleteUser(uuid); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusInternalServerError,
						Title:   "Internal server error",
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

// @Summary GetUser
// @Tags User
// @Description Get user
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Router /api/v1/users/{id} [get]
// @Security ApiKeyAuth
func (h *userHandler) GetUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		uuid, err := uuid.Parse(id)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Title:   "Invalid user id",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}

		user, err := h.userService.GetUser(uuid)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusInternalServerError,
						Title:   "Internal server error",
						Message: err.Error(),
						Source:  helpers.WhereAmI(),
					},
				},
			})
		}

		return c.Status(fiber.StatusOK).JSON(helpers.ResponseForm{
			Success: true,
			Data:    user,
		})
	}
}
