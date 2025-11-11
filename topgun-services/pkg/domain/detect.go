package domain

import "topgun-services/pkg/models"

type DetectRepository interface {
	CreateDetect(detect models.Detect) (*models.Detect, error)
	GetDetects(pagination models.Pagination, filter models.Search) ([]models.Detect, *models.Pagination, *models.Search, error)
	GetDetect(id uint) (*models.Detect, error)
	GetDetectFile(id uint) (*models.Detect, error)
	UpdateDetect(id uint, detect models.Detect) (*models.Detect, error)
	DeleteDetect(id uint) error
}
type DetectService interface {
	CreateDetect(detect models.Detect) (*models.Detect, error)
	GetDetects(pagination models.Pagination, filter models.Search) ([]models.Detect, *models.Pagination, *models.Search, error)
	GetDetect(id uint) (*models.Detect, error)
	GetDetectFile(id uint) (*models.Detect, error)
	UpdateDetect(id uint, detect models.Detect) (*models.Detect, error)
	DeleteDetect(id uint) error
}
