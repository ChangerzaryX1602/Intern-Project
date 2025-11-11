package user

import (
	"topgun-services/pkg/domain"
	"topgun-services/pkg/models"

	"github.com/google/uuid"
)

type userService struct {
	domain.UserRepository
}

func NewUserService(userRepo domain.UserRepository) domain.UserService {
	return &userService{
		UserRepository: userRepo,
	}
}
func (s *userService) GetUsers(pagination models.Pagination, filter models.Search) ([]models.User, *models.Pagination, *models.Search, error) {
	return s.UserRepository.GetUsers(pagination, filter)
}
func (s *userService) CreateUser(user models.User) (*models.User, error) {
	return s.UserRepository.CreateUser(user)
}
func (s *userService) UpdateUser(id uuid.UUID, user models.User) (*models.User, error) {
	return s.UserRepository.UpdateUser(id, user)
}
func (s *userService) DeleteUser(id uuid.UUID) error {
	return s.UserRepository.DeleteUser(id)
}
func (s *userService) GetUser(id uuid.UUID) (*models.User, error) {
	return s.UserRepository.GetUser(id)
}
func (s *userService) GetUserByEmail(email string) (*models.User, error) {
	return s.UserRepository.GetUserByEmail(email)
}

func (s *userService) UpdateUserPassword(id uuid.UUID, newPassword string) error {
	return s.UserRepository.UpdateUserPassword(id, newPassword)
}
