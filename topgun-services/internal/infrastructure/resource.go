package infrastructure

import (
	"fmt"

	"topgun-services/internal/datasources"
	"topgun-services/pkg/models"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gofiber/storage/redis"
	"github.com/golang-jwt/jwt/v4"
	"github.com/valyala/fasthttp"
	"gorm.io/gorm"
)

func NewResources(fasthttpClient *fasthttp.Client, mainDbConn *gorm.DB, logDbConn *gorm.DB, redisStorage *redis.Storage, jwtResources *models.JwtResources, mqttClient mqtt.Client) models.Resources {
	return models.Resources{
		FastHTTPClient: fasthttpClient,
		MainDbConn:     mainDbConn,
		LogDbConn:      logDbConn,
		RedisStorage:   redisStorage,
		JwtResources:   jwtResources,
		MQTTClient:     mqttClient,
	}
}

func NewJwt(privateKeyPath string) (jwtResources *models.JwtResources, err error) {
	jwtResources = new(models.JwtResources)
	jwtResources.JwtSignKey, jwtResources.JwtVerifyKey, jwtResources.JwtSigningMethod, err = datasources.NewJwtLocalKey(privateKeyPath)
	jwtResources.JwtKeyfunc = func(token *jwt.Token) (publicKey interface{}, err error) {
		if jwtResources.JwtVerifyKey == nil {
			err = fmt.Errorf("JWTVerifyKey not init yet")
		}
		return jwtResources.JwtVerifyKey, err
	}
	jwtResources.JwtParser = jwt.NewParser()
	return
}
