package detect

import (
	"topgun-services/pkg/domain"
	"topgun-services/pkg/models"
	"topgun-services/pkg/utils"

	"gorm.io/gorm"
)

type detectRepository struct {
	DB *gorm.DB
}

func NewDetectRepository(db *gorm.DB) domain.DetectRepository {
	return &detectRepository{DB: db}
}
func (r *detectRepository) CreateDetect(detect models.Detect) (*models.Detect, error) {
	if r.DB == nil {
		return nil, gorm.ErrInvalidDB
	}
	err := r.DB.Create(&detect).Error
	if err != nil {
		return nil, err
	}
	return &detect, nil
}
func (r *detectRepository) GetDetects(pagination models.Pagination, filter models.Search) ([]models.Detect, *models.Pagination, *models.Search, error) {
	if r.DB == nil {
		return nil, nil, nil, gorm.ErrInvalidDB
	}
	var detects []models.Detect
	dbTx := utils.ApplySearch(r.DB, filter)
	dbTx = utils.ApplyPagination(dbTx, &pagination, &detects)
	err := dbTx.Find(&detects).Error
	if err != nil {
		return nil, nil, nil, err
	}
	return detects, &pagination, &filter, nil
}

func (r *detectRepository) GetDetectsByCameras(cameraIDs []string, pagination models.Pagination) ([]models.Detect, *models.Pagination, error) {
	if r.DB == nil {
		return nil, nil, gorm.ErrInvalidDB
	}
	var detects []models.Detect

	// Query detections where camera_id is in the provided list
	dbTx := r.DB.Where("camera_id IN ?", cameraIDs).Order("timestamp DESC")
	dbTx = utils.ApplyPagination(dbTx, &pagination, &detects)

	err := dbTx.Find(&detects).Error
	if err != nil {
		return nil, nil, err
	}

	return detects, &pagination, nil
}

func (r *detectRepository) GetDetect(id uint) (*models.Detect, error) {
	if r.DB == nil {
		return nil, gorm.ErrInvalidDB
	}
	var detect models.Detect
	err := r.DB.First(&detect, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &detect, nil
}
func (r *detectRepository) GetDetectFile(id uint) (*models.Detect, error) {
	if r.DB == nil {
		return nil, gorm.ErrInvalidDB
	}
	var detect models.Detect
	err := r.DB.Select("id, path").First(&detect, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &detect, nil
}
func (r *detectRepository) UpdateDetect(id uint, detect models.Detect) (*models.Detect, error) {
	if r.DB == nil {
		return nil, gorm.ErrInvalidDB
	}
	var existingDetect models.Detect
	err := r.DB.First(&existingDetect, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	err = r.DB.Model(&existingDetect).Updates(detect).Error
	if err != nil {
		return nil, err
	}
	return &existingDetect, nil
}
func (r *detectRepository) DeleteDetect(id uint) error {
	if r.DB == nil {
		return gorm.ErrInvalidDB
	}
	err := r.DB.Where("id = ?", id).First(&models.Detect{}).Error
	if err != nil {
		return err
	}
	err = r.DB.Delete(&models.Detect{}, "id = ?", id).Error
	if err != nil {
		return err
	}
	return nil
}
