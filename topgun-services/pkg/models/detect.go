package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// JSONRawMessageArray is a custom type to handle JSONB array scanning
type JSONRawMessageArray []json.RawMessage

// Scan implements the sql.Scanner interface
func (j *JSONRawMessageArray) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal JSONB value")
	}

	// Try to unmarshal as array first
	var arr []json.RawMessage
	if err := json.Unmarshal(bytes, &arr); err == nil {
		*j = arr
		return nil
	}

	// If it's not an array, try to unmarshal as single object and wrap it in array
	var obj json.RawMessage
	if err := json.Unmarshal(bytes, &obj); err != nil {
		return errors.New("failed to unmarshal JSONB value as array or object")
	}
	*j = []json.RawMessage{obj}
	return nil
}

// Value implements the driver.Valuer interface
func (j JSONRawMessageArray) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

type Detect struct {
	ID        uint                `gorm:"primaryKey;autoIncrement" json:"id"`
	CameraID  uuid.UUID           `json:"camera_id"`
	Timestamp time.Time           `json:"timestamp" gorm:"autoCreateTime;default:CURRENT_TIMESTAMP" swaggerignore:"true"`
	Camera    Camera              `gorm:"foreignKey:CameraID;references:ID" json:"camera"`
	Path      string              `json:"path"`
	Objects   JSONRawMessageArray `json:"objects" gorm:"type:jsonb" swaggerignore:"true"`
}
