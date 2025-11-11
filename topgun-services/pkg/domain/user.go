package domain

import (
	"topgun-services/pkg/models"

	"github.com/google/uuid"
)

type UserRepository interface {
	GetUsers(pagination models.Pagination, filter models.Search) ([]models.User, *models.Pagination, *models.Search, error)
	CreateUser(user models.User) (*models.User, error)
	UpdateUser(id uuid.UUID, user models.User) (*models.User, error)
	UpdateUserPassword(id uuid.UUID, newPassword string) error
	DeleteUser(id uuid.UUID) error
	GetUser(id uuid.UUID) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
}
type UserService interface {
	GetUsers(pagination models.Pagination, filter models.Search) ([]models.User, *models.Pagination, *models.Search, error)
	CreateUser(user models.User) (*models.User, error)
	UpdateUser(id uuid.UUID, user models.User) (*models.User, error)
	UpdateUserPassword(id uuid.UUID, newPassword string) error
	DeleteUser(id uuid.UUID) error
	GetUser(id uuid.UUID) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
}
