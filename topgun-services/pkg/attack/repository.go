package attack

import (
	"topgun-services/pkg/domain"
	"topgun-services/pkg/models"
	"topgun-services/pkg/utils"

	"gorm.io/gorm"
)

type attackRepository struct {
	DB *gorm.DB
}

func NewAttackRepository(db *gorm.DB) domain.AttackRepository {
	return &attackRepository{DB: db}
}
func (r *attackRepository) GetAttacks(pagination models.Pagination, filter models.Search) ([]models.Attack, *models.Pagination, *models.Search, error) {
	if r.DB == nil {
		return nil, nil, nil, gorm.ErrInvalidDB
	}
	var attacks []models.Attack
	dbTx := r.DB.Model(&models.Attack{})
	dbTx = utils.ApplySearch(dbTx, filter)
	dbTx = utils.ApplyPagination(dbTx, &pagination, &attacks)
	err := dbTx.Find(&attacks).Error
	return attacks, &pagination, &filter, err
}
func (r *attackRepository) CreateAttack(attack models.Attack) (*models.Attack, error) {
	if r.DB == nil {
		return nil, gorm.ErrInvalidDB
	}
	err := r.DB.Create(&attack).Error
	if err != nil {
		return nil, err
	}
	return &attack, nil
}
func (r *attackRepository) UpdateAttack(id uint, attack models.Attack) (*models.Attack, error) {
	if r.DB == nil {
		return nil, gorm.ErrInvalidDB
	}
	var existing models.Attack
	err := r.DB.First(&existing, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	err = r.DB.Model(&existing).Updates(attack).Error
	if err != nil {
		return nil, err
	}
	return &existing, nil
}
func (r *attackRepository) DeleteAttack(id uint) error {
	if r.DB == nil {
		return gorm.ErrInvalidDB
	}
	err := r.DB.Delete(&models.Attack{}, "id = ?", id).Error
	return err
}
func (r *attackRepository) GetAttack(id uint) (*models.Attack, error) {
	if r.DB == nil {
		return nil, gorm.ErrInvalidDB
	}
	var attack models.Attack
	err := r.DB.First(&attack, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &attack, nil
}
