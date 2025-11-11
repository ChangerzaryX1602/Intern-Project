package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Camera struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string    `json:"name"`
	Location  string    `json:"location"`
	Token     string    `json:"token"`
	Institute string    `json:"institute"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime;default:CURRENT_TIMESTAMP" swaggerignore:"true"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime;default:CURRENT_TIMESTAMP" swaggerignore:"true"`
}

func (u *Camera) BeforeCreate(tx *gorm.DB) error {
	u.ID = uuid.New()
	return nil
}
