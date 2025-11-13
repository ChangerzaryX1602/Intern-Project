package detect

import (
	"time"
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
func (r *detectRepository) GetDetects(pagination models.Pagination, filter models.Search, startDate, endDate string) ([]models.Detect, *models.Pagination, *models.Search, error) {
	if r.DB == nil {
		return nil, nil, nil, gorm.ErrInvalidDB
	}
	var detects []models.Detect
	dbTx := utils.ApplySearch(r.DB, filter)

	// Apply date range filter if provided
	if startDate != "" {
		// Parse start date (format: YYYY-MM-DD)
		if startTime, err := time.Parse("2006-01-02", startDate); err == nil {
			// Set to beginning of day
			startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, startTime.Location())
			dbTx = dbTx.Where("timestamp >= ?", startTime)
		}
	}

	if endDate != "" {
		// Parse end date (format: YYYY-MM-DD)
		if endTime, err := time.Parse("2006-01-02", endDate); err == nil {
			// Set to end of day (23:59:59)
			endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 999999999, endTime.Location())
			dbTx = dbTx.Where("timestamp <= ?", endTime)
		}
	}

	// Order by timestamp descending (newest first)
	dbTx = dbTx.Order("timestamp DESC")

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
