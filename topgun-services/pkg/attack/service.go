package attack

import (
	"topgun-services/pkg/domain"
	"topgun-services/pkg/models"
)

type attackService struct {
	repository domain.AttackRepository
}

func NewAttackService(repo domain.AttackRepository) domain.AttackService {
	return &attackService{repository: repo}
}
func (s *attackService) GetAttacks(pagination models.Pagination, filter models.Search) ([]models.Attack, *models.Pagination, *models.Search, error) {
	return s.repository.GetAttacks(pagination, filter)
}
func (s *attackService) CreateAttack(attack models.Attack) (*models.Attack, error) {
	return s.repository.CreateAttack(attack)
}
func (s *attackService) UpdateAttack(id uint, attack models.Attack) (*models.Attack, error) {
	return s.repository.UpdateAttack(id, attack)
}
func (s *attackService) DeleteAttack(id uint) error {
	return s.repository.DeleteAttack(id)
}
func (s *attackService) GetAttack(id uint) (*models.Attack, error) {
	return s.repository.GetAttack(id)
}
