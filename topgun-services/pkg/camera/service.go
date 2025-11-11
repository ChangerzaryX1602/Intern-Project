package camera

import (
	"topgun-services/pkg/domain"
	"topgun-services/pkg/models"

	"github.com/google/uuid"
)

type cameraService struct {
	repository domain.CameraRepository
}

func NewCameraService(repo domain.CameraRepository) domain.CameraService {
	return &cameraService{repository: repo}
}
func (s *cameraService) GetCameras(pagination models.Pagination, filter models.Search) ([]models.Camera, *models.Pagination, *models.Search, error) {
	return s.repository.GetCameras(pagination, filter)
}
func (s *cameraService) CreateCamera(camera models.Camera) (*models.Camera, error) {
	return s.repository.CreateCamera(camera)
}
func (s *cameraService) UpdateCamera(id uuid.UUID, camera models.Camera) (*models.Camera, error) {
	return s.repository.UpdateCamera(id, camera)
}
func (s *cameraService) DeleteCamera(id uuid.UUID) error {
	return s.repository.DeleteCamera(id)
}
func (s *cameraService) GetCamera(id uuid.UUID) (*models.Camera, error) {
	return s.repository.GetCamera(id)
}
