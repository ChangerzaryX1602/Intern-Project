package camera

import (
	"topgun-services/pkg/domain"
	"topgun-services/pkg/models"
	"topgun-services/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type cameraRepository struct {
	DB *gorm.DB
}

func NewCameraRepository(db *gorm.DB) domain.CameraRepository {
	return &cameraRepository{DB: db}
}
func (r *cameraRepository) GetCameras(pagination models.Pagination, filter models.Search) ([]models.Camera, *models.Pagination, *models.Search, error) {
	if r.DB == nil {
		return nil, nil, nil, gorm.ErrInvalidDB
	}
	var cameras []models.Camera
	dbTx := utils.ApplySearch(r.DB, filter)
	dbTx = utils.ApplyPagination(dbTx, &pagination, &cameras)
	err := dbTx.Find(&cameras).Error
	if err != nil {
		return nil, nil, nil, err
	}
	return cameras, &pagination, &filter, nil
}
func (r *cameraRepository) CreateCamera(camera models.Camera) (*models.Camera, error) {
	if r.DB == nil {
		return nil, gorm.ErrInvalidDB
	}
	err := r.DB.Create(&camera).Error
	if err != nil {
		return nil, err
	}
	return &camera, nil
}
func (r *cameraRepository) UpdateCamera(id uuid.UUID, camera models.Camera) (*models.Camera, error) {
	if r.DB == nil {
		return nil, gorm.ErrInvalidDB
	}
	var existingCamera models.Camera
	err := r.DB.First(&existingCamera, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	err = r.DB.Model(&existingCamera).Updates(camera).Error
	if err != nil {
		return nil, err
	}
	return &existingCamera, nil
}
func (r *cameraRepository) DeleteCamera(id uuid.UUID) error {
	if r.DB == nil {
		return gorm.ErrInvalidDB
	}
	err := r.DB.Where("id = ?", id).First(&models.Camera{}).Error
	if err != nil {
		return err
	}
	err = r.DB.Delete(&models.Camera{}, "id = ?", id).Error
	if err != nil {
		return err
	}
	return nil
}
func (r *cameraRepository) GetCamera(id uuid.UUID) (*models.Camera, error) {
	if r.DB == nil {
		return nil, gorm.ErrInvalidDB
	}
	var camera models.Camera
	err := r.DB.First(&camera, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &camera, nil
}
