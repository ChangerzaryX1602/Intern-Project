package domain

import (
	"topgun-services/pkg/models"

	"github.com/google/uuid"
)

type CameraRepository interface {
	GetCameras(pagination models.Pagination, filter models.Search) ([]models.Camera, *models.Pagination, *models.Search, error)
	CreateCamera(camera models.Camera) (*models.Camera, error)
	UpdateCamera(id uuid.UUID, camera models.Camera) (*models.Camera, error)
	DeleteCamera(id uuid.UUID) error
	GetCamera(id uuid.UUID) (*models.Camera, error)
}
type CameraService interface {
	GetCameras(pagination models.Pagination, filter models.Search) ([]models.Camera, *models.Pagination, *models.Search, error)
	CreateCamera(camera models.Camera) (*models.Camera, error)
	UpdateCamera(id uuid.UUID, camera models.Camera) (*models.Camera, error)
	DeleteCamera(id uuid.UUID) error
	GetCamera(id uuid.UUID) (*models.Camera, error)
}
