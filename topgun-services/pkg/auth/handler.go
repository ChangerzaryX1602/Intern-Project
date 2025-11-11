package auth

import (
	"fmt"

	"topgun-services/internal/handlers"
	"topgun-services/pkg/domain"
	"topgun-services/pkg/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	helpers "github.com/zercle/gofiber-helpers"
)

type authHandler struct {
	authService domain.AuthService
	userService domain.UserService
}

func NewAuthHandler(router fiber.Router, routerResource *handlers.RouterResources, authService domain.AuthService, userService domain.UserService) {
	handler := &authHandler{
		authService: authService,
		userService: userService,
	}
	router.Post("/register", handler.Register())
	router.Post("/login", handler.Login())
	router.Post("/reset-password", routerResource.ReqAuthHandler(), handler.ResetPassword())
	router.Post("/reset-password/verify", handler.VerifyTokenAndResetPassword())
	router.Post("/reset-password/:email/issue-token", handler.IssueTokenVerification())
}

// IssueTokenVerification godoc
// @Summary Issue token for email verification
// @Description Issue token for email verification
// @Tags Auth
// @Accept json
// @Produce json
// @Param email path string true "Email"
// @Router /api/v1/auth/reset-password/{email}/issue-token [post]
func (h *authHandler) IssueTokenVerification() fiber.Handler {
	return func(c *fiber.Ctx) error {
		email := c.Params("email")
		scheme := "http"
		if c.Protocol() == "https" {
			scheme = "https"
		}
		hostWithScheme := fmt.Sprintf("%s://%s", scheme, c.Hostname())
		_, err := h.authService.IssueTokenVerification(email, hostWithScheme)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusInternalServerError,
						Source:  helpers.WhereAmI(),
						Title:   "Internal Server Error",
						Message: err.Error(),
					},
				},
			})
		}
		return c.Status(fiber.StatusOK).JSON(helpers.ResponseForm{
			Success: true,
		})
	}
}

// VerifyTokenAndResetPassword godoc
// @Summary Verify token and reset password
// @Description Verify token and reset password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.VerifyTokenAndResetPasswordRequest true "VerifyTokenAndResetPasswordRequest"
// @Router /api/v1/auth/reset-password/verify [post]
func (h *authHandler) VerifyTokenAndResetPassword() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.VerifyTokenAndResetPasswordRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Source:  helpers.WhereAmI(),
						Title:   "Bad Request",
						Message: err.Error(),
					},
				},
			})
		}
		if err := h.authService.VerifyTokenAndResetPassword(req.Token, req.Password); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusUnauthorized,
						Source:  helpers.WhereAmI(),
						Title:   "Unauthorized",
						Message: err.Error(),
					},
				},
			})
		}
		return c.Status(fiber.StatusOK).JSON(helpers.ResponseForm{
			Success: true,
		})
	}
}

// @Summary Register
// @Description Register a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body models.User true "User"
// @Router /api/v1/auth/register [post]
func (h *authHandler) Register() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var user models.User
		if err := c.BodyParser(&user); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Source:  helpers.WhereAmI(),
						Title:   "Bad Request",
						Message: err.Error(),
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
						Source:  helpers.WhereAmI(),
						Title:   "Internal Server Error",
						Message: err.Error(),
					},
				},
			})
		}
		return c.Status(fiber.StatusCreated).JSON(helpers.ResponseForm{
			Success: true,
			Data:    map[string]interface{}{"user": u},
		})
	}
}

// @Summary Login
// @Description Login a user
// @Tags Auth
// @Accept json
// @Produce json
// @Param login body models.LoginRequest true "Login Request"
// @Router /api/v1/auth/login [post]
func (h *authHandler) Login() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Source:  helpers.WhereAmI(),
						Title:   "Bad Request",
						Message: err.Error(),
					},
				},
			})
		}
		user, err := h.userService.GetUserByEmail(req.Email)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusUnauthorized,
						Source:  helpers.WhereAmI(),
						Title:   "Unauthorized",
						Message: "Invalid email",
					},
				},
			})
		}
		if !user.CheckPassword(req.Password) {
			return c.Status(fiber.StatusUnauthorized).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusUnauthorized,
						Source:  helpers.WhereAmI(),
						Title:   "Unauthorized",
						Message: "Invalid password",
					},
				},
			})
		}
		token, err := h.authService.SignToken(user, c.Hostname())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusInternalServerError,
						Source:  helpers.WhereAmI(),
						Title:   "Internal Server Error",
						Message: err.Error(),
					},
				},
			})
		}
		return c.Status(fiber.StatusOK).JSON(helpers.ResponseForm{
			Success: true,
			Data: fiber.Map{
				"token": token,
				"user":  user,
			},
		})
	}
}

// @Summary Reset Password
// @Description Reset user password with old and new password
// @Tags Auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param resetPassword body models.ResetPasswordRequest true "Reset Password Request"
// @Router /api/v1/auth/reset-password [post]
func (h *authHandler) ResetPassword() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := uuid.Parse(c.Locals("user_id").(string))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Source:  helpers.WhereAmI(),
						Title:   "Bad Request",
						Message: "Invalid user ID",
					},
				},
			})
		}

		var req models.ResetPasswordRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors: []helpers.ResponseError{
					{
						Code:    fiber.StatusBadRequest,
						Source:  helpers.WhereAmI(),
						Title:   "Bad Request",
						Message: err.Error(),
					},
				},
			})
		}

		errs := h.authService.ResetPassword(userID, &req)
		if errs != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helpers.ResponseForm{
				Success: false,
				Errors:  errs,
			})
		}

		return c.Status(fiber.StatusOK).JSON(helpers.ResponseForm{
			Success: true,
			Data: fiber.Map{
				"message": "Password reset successfully",
			},
		})
	}
}
