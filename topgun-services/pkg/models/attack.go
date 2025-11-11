package models

import "time"

type Attack struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Lat           float32   `json:"lat"`
	Lng           float32   `json:"lng"`
	Height        float32   `json:"height"`
	Function      string    `json:"function"`
	Accelleration float32   `json:"acceleration"`
	Velocity      float32   `json:"velocity"`
	Distance      float32   `json:"distance"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime;default:CURRENT_TIMESTAMP" swaggerignore:"true"`
}
