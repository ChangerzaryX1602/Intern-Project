package models

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey" swaggerignore:"true"`
	Title        string         `json:"title"`
	TitleEn      string         `json:"title_en"`
	FirstName    string         `json:"first_name"`
	FirstNameEn  string         `json:"first_name_en"`
	LastName     string         `json:"last_name"`
	LastNameEn   string         `json:"last_name_en"`
	Email        string         `json:"email"`
	Password     string         `json:"-"`
	PasswordTemp string         `json:"password" gorm:"-"`
	CreatedAt    time.Time      `json:"created_at" gorm:"autoCreateTime;default:CURRENT_TIMESTAMP" swaggerignore:"true"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime;default:CURRENT_TIMESTAMP" swaggerignore:"true"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index;default:null" swaggerignore:"true"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.ID = uuid.New()
	return nil
}
func (u *User) BeforeSave(tx *gorm.DB) error {
	if u.PasswordTemp != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.PasswordTemp), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Password = string(hashedPassword)
		u.PasswordTemp = ""
	}
	return nil
}
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
