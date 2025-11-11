package domain

import "topgun-services/pkg/models"

type AttackRepository interface {
	GetAttacks(pagination models.Pagination, filter models.Search) ([]models.Attack, *models.Pagination, *models.Search, error)
	CreateAttack(attack models.Attack) (*models.Attack, error)
	UpdateAttack(id uint, attack models.Attack) (*models.Attack, error)
	DeleteAttack(id uint) error
	GetAttack(id uint) (*models.Attack, error)
}
type AttackService interface {
	GetAttacks(pagination models.Pagination, filter models.Search) ([]models.Attack, *models.Pagination, *models.Search, error)
	CreateAttack(attack models.Attack) (*models.Attack, error)
	UpdateAttack(id uint, attack models.Attack) (*models.Attack, error)
	DeleteAttack(id uint) error
	GetAttack(id uint) (*models.Attack, error)
}
