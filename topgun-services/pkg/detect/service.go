package detect

import (
	"topgun-services/pkg/domain"
	"topgun-services/pkg/models"
)

type detectService struct {
	repository domain.DetectRepository
}

func NewDetectService(repo domain.DetectRepository) domain.DetectService {
	return &detectService{repository: repo}
}
func (s *detectService) CreateDetect(detect models.Detect) (*models.Detect, error) {
	return s.repository.CreateDetect(detect)
}
func (s *detectService) GetDetects(pagination models.Pagination, filter models.Search) ([]models.Detect, *models.Pagination, *models.Search, error) {
	return s.repository.GetDetects(pagination, filter)
}
func (s *detectService) GetDetectsByCameras(cameraIDs []string, pagination models.Pagination) ([]models.Detect, *models.Pagination, error) {
	return s.repository.GetDetectsByCameras(cameraIDs, pagination)
}
func (s *detectService) GetDetect(id uint) (*models.Detect, error) {
	return s.repository.GetDetect(id)
}
func (s *detectService) GetDetectFile(id uint) (*models.Detect, error) {
	return s.repository.GetDetectFile(id)
}
func (s *detectService) UpdateDetect(id uint, detect models.Detect) (*models.Detect, error) {
	return s.repository.UpdateDetect(id, detect)
}
func (s *detectService) DeleteDetect(id uint) error {
	return s.repository.DeleteDetect(id)
}
