package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"topgun-services/pkg/domain"
	"topgun-services/pkg/models"

	"github.com/gofiber/storage/redis"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type authRepository struct {
	*gorm.DB
	*models.JwtResources
	*redis.Storage
}

func NewAuthRepository(db *gorm.DB, jwt *models.JwtResources, redis *redis.Storage) domain.AuthRepository {
	return &authRepository{
		DB:           db,
		JwtResources: jwt,
		Storage:      redis,
	}
}

func (r *authRepository) SignToken(user *models.User, host string) (string, error) {
	token := jwt.NewWithClaims(r.JwtSigningMethod, &jwt.RegisteredClaims{})
	claims := token.Claims.(*jwt.RegisteredClaims)
	claims.Subject = user.ID.String()
	claims.Issuer = host
	claims.Audience = []string{"permission:10"}
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Hour * 24))
	signToken, err := token.SignedString(r.JwtSignKey)
	if err != nil {
		return "", err
	}
	return signToken, nil
}

func (r *authRepository) IssueTokenVerification(email string) (*string, error) {
	if r.DB == nil {
		return nil, errors.New("database not available")
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	token := base64.RawURLEncoding.EncodeToString(b)
	sum := sha256.Sum256([]byte(token))
	tokenHash := base64.RawURLEncoding.EncodeToString(sum[:])
	key := "reset:" + tokenHash
	if err := r.Storage.Set(key, []byte(email), 10*time.Minute); err != nil {
		return nil, err
	}
	return &token, nil
}
func (r *authRepository) VerifyToken(token string) (uuid.UUID, error) {
	if r.DB == nil {
		return uuid.Nil, errors.New("database not available")
	}
	sum := sha256.Sum256([]byte(token))
	tokenHash := base64.RawURLEncoding.EncodeToString(sum[:])
	key := "reset:" + tokenHash
	email, err := r.Storage.Get(key)
	if err != nil {
		return uuid.Nil, err
	}
	if email == nil {
		return uuid.Nil, errors.New("token not found or expired")
	}
	var user models.User
	emailString := string(email)
	emailString = strings.ReplaceAll(emailString, "%40", "@")
	err = r.DB.Where("email = ?", emailString).First(&user).Error
	if err != nil {
		return uuid.Nil, err
	}
	if err := r.Storage.Delete(key); err != nil {
		return uuid.Nil, err
	}
	return user.ID, nil

}
