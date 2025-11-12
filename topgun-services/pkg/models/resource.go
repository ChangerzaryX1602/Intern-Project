package models

import (
	"crypto"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/redis"
	"github.com/golang-jwt/jwt/v4"
	"github.com/valyala/fasthttp"
	"gorm.io/gorm"
)

type Resources struct {
	LogConfig      logger.Config
	FastHTTPClient *fasthttp.Client
	MainDbConn     *gorm.DB
	LogDbConn      *gorm.DB
	RedisStorage   *redis.Storage
	JwtResources   *JwtResources
	SessConfig     session.Config
	MQTTClient     mqtt.Client
}

type JwtResources struct {
	JwtVerifyKey     crypto.PublicKey
	JwtSignKey       crypto.PrivateKey
	JwtSigningMethod jwt.SigningMethod
	JwtKeyfunc       jwt.Keyfunc
	JwtParser        *jwt.Parser
}
