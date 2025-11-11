package handlers

import (
	"github.com/gofiber/storage/redis"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

// RouterResources DB handler
type RouterResources struct {
	JwtKeyfunc jwt.Keyfunc
	DB         *gorm.DB
	Store      *redis.Storage
}

// NewRouterResources returns a new DBHandler
func NewRouterResources(jwtKeyfunc jwt.Keyfunc, DB *gorm.DB, Redis *redis.Storage) *RouterResources {
	return &RouterResources{
		JwtKeyfunc: jwtKeyfunc,
		DB:         DB,
		Store:      Redis,
	}
}
