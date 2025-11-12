package models

import "time"

type Attack struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Acceleration Vector    `json:"acceleration" gorm:"type:jsonb"`
	Distance     float32   `json:"distance"`
	DroneID      string    `json:"drone_id"`
	Height       float32   `json:"height"`
	Lat          float32   `json:"lat"`
	Lng          float32   `json:"lng"`
	Status       string    `json:"status"`
	Velocity     Vector    `json:"velocity" gorm:"type:jsonb"`
	TimeLeft     int       `json:"time_left"`
	Target       Location  `json:"target" gorm:"type:jsonb"`
	Landing      Location  `json:"landing" gorm:"type:jsonb"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime;default:CURRENT_TIMESTAMP" swaggerignore:"true"`
}
type AttackRequest struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Acceleration Vector    `json:"acceleration" gorm:"type:jsonb"`
	Distance     float32   `json:"distance"`
	DroneID      string    `json:"drone_id"`
	Height       float32   `json:"height"`
	Lat          float32   `json:"lat"`
	Lng          float32   `json:"lng"`
	Status       string    `json:"status"`
	Velocity     Vector    `json:"velocity" gorm:"type:jsonb"`
	TimeLeft     float32   `json:"time_left"`
	Target       Location  `json:"target" gorm:"type:jsonb"`
	Landing      Location  `json:"landing" gorm:"type:jsonb"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime;default:CURRENT_TIMESTAMP" swaggerignore:"true"`
}
type Vector struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
}
type Location struct {
	Lat         float32 `json:"lat"`
	Lng         float32 `json:"lng"`
	Description string  `json:"description"`
}
