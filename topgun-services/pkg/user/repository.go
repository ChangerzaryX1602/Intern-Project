package user

import (
	"errors"
	"net/http"

	"topgun-services/pkg/domain"
	"topgun-services/pkg/models"
	"topgun-services/pkg/utils"

	"github.com/google/uuid"
	helpers "github.com/zercle/gofiber-helpers"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type userRepository struct {
	*gorm.DB
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db}
}
func (r *userRepository) GetUsers(pagination models.Pagination, filter models.Search) ([]models.User, *models.Pagination, *models.Search, error) {
	if r.DB == nil {
		return nil, nil, nil, gorm.ErrInvalidDB
	}
	var users []models.User
	dbTx := utils.ApplySearch(r.DB, filter)
	dbTx = utils.ApplyPagination(dbTx, &pagination, &users)
	err := dbTx.Find(&users).Error
	if err != nil {
		return nil, nil, nil, err
	}
	return users, &pagination, &filter, nil
}
func (r *userRepository) CreateUser(user models.User) (*models.User, error) {
	if r.DB == nil {
		return nil, gorm.ErrInvalidDB
	}
	if user.Email != "" {
		existingUserEmail, err := r.GetUserByEmail(user.Email)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if existingUserEmail != nil {
			return nil, helpers.NewError(http.StatusConflict, "email is already in use by another account")
		}
	}
	err := r.DB.Create(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func (r *userRepository) UpdateUser(id uuid.UUID, user models.User) (*models.User, error) {
	if r.DB == nil {
		return nil, gorm.ErrInvalidDB
	}
	var newUser models.User
	err := r.DB.Model(&newUser).Where("id = ?", id).Updates(user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func (r *userRepository) DeleteUser(id uuid.UUID) error {
	if r.DB == nil {
		return gorm.ErrInvalidDB
	}
	err := r.DB.Where("id = ?", id).First(&models.User{}).Error
	if err != nil {
		return err
	}
	err = r.DB.Where("id = ?", id).Delete(&models.User{}).Error
	if err != nil {
		return err
	}
	return nil
}
func (r *userRepository) GetUser(id uuid.UUID) (*models.User, error) {
	if r.DB == nil {
		return nil, gorm.ErrInvalidDB
	}
	var user models.User
	err := r.DB.Where("id = ?", id).Preload("UserRoles.Role").Preload("UserRoles.Faculty").First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func (r *userRepository) GetUserByEmail(email string) (*models.User, error) {
	if r.DB == nil {
		return nil, gorm.ErrInvalidDB
	}
	var user models.User
	err := r.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func (r *userRepository) GetUserByPID(pid string) (*models.User, error) {
	if r.DB == nil {
		return nil, gorm.ErrInvalidDB
	}

	var user models.User
	if err := r.DB.Where("pid = ?", pid).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) UpdateUserPassword(id uuid.UUID, newPassword string) error {
	if r.DB == nil {
		return gorm.ErrInvalidDB
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := models.User{
		Password: string(hashedPassword),
	}

	err = r.DB.Model(&models.User{}).Where("id = ?", id).Updates(&user).Error
	return err
}
