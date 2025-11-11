package models

import (
	"time"

	"github.com/google/uuid"
)

type Log struct {
	ID            uint      `gorm:"primaryKey"`
	At            time.Time `gorm:"index"`
	Status        int
	IP            string
	Method        string
	Host          string
	URL           string    `gorm:"type:text"`
	UserAgent     string    `gorm:"type:text"`
	UserID        uuid.UUID `gorm:"index"`
	Referer       string    `gorm:"type:text"`
	Authorization string    `gorm:"type:text"`
	BytesRecv     int
	BytesSent     int
	ErrorMsg      string `gorm:"type:text"`
}
