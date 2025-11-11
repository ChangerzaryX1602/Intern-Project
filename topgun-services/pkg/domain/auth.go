package domain

import (
	"topgun-services/pkg/models"

	"github.com/google/uuid"
	helpers "github.com/zercle/gofiber-helpers"
)

type AuthRepository interface {
	SignToken(user *models.User, host string) (string, error)
	IssueTokenVerification(email string) (*string, error)
	VerifyToken(token string) (uuid.UUID, error)
	// Firebase methods
}
type AuthService interface {
	SignToken(user *models.User, host string) (string, error)
	ResetPassword(userID uuid.UUID, body *models.ResetPasswordRequest) []helpers.ResponseError
	VerifyTokenAndResetPassword(token string, password string) error
	IssueTokenVerification(email string, host string) (*string, error)
}
