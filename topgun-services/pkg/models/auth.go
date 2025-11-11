package models

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type VerifyTokenAndResetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type ThaIDClaims struct {
	PID       string `json:"pid"`
	Birthdate string `json:"birthdate"`
	Address   struct {
		Formatted string `json:"formatted"`
	} `json:"address"`
	GivenName    string `json:"given_name"`
	GivenNameEn  string `json:"given_name_en"`
	FamilyName   string `json:"family_name"`
	FamilyNameEn string `json:"family_name_en"`
	MiddleName   string `json:"middle_name"`
	MiddleNameEn string `json:"middle_name_en"`
	Gender       string `json:"gender"`
	TitleTh      string `json:"titleTh"`
	TitleEn      string `json:"titleEn"`
	HouseAddress struct {
		Formatted string `json:"formatted"`
		Raw       string `json:"raw"`
	} `json:"house_address"`

	jwt.RegisteredClaims
}

type LoginThaIDRequest struct {
	RedirectURI string `json:"redirect_uri"`
	Code        string `json:"code"`
}

type EmailToken struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Token     string    `gorm:"unique;not null" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expiresAt"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

type RequestEmailVerify struct {
	Email       string `json:"email"`
	RedirectURL string `json:"redirectURL"`
}

type RequestEmailToken struct {
	Token string `json:"token"`
}

type ResetPasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

type Oauth struct {
	Code string `json:"code"`
}

// Firebase OAuth requests
type FirebaseOAuthRequest struct {
	IDToken string `json:"id_token" validate:"required"`
}

type GoogleOAuthRequest struct {
	IDToken string `json:"id_token" validate:"required"`
}

type MicrosoftOAuthRequest struct {
	IDToken string `json:"id_token" validate:"required"`
}

type FirebaseUserInfo struct {
	UID         string `json:"uid"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	PhotoURL    string `json:"photo_url"`
	Provider    string `json:"provider"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
}
