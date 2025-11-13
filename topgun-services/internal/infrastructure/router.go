package infrastructure

import (
	"fmt"
	"log"
	"time"

	"topgun-services/internal/handlers"
	"topgun-services/pkg/attack"
	"topgun-services/pkg/auth"
	"topgun-services/pkg/camera"
	"topgun-services/pkg/detect"
	"topgun-services/pkg/logs"
	"topgun-services/pkg/models"
	"topgun-services/pkg/mqtt"
	"topgun-services/pkg/user"

	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

// SetupRoutes is the Router for GoFiber App
func (s *Server) SetupRoutes(app *fiber.App) {
	// Prepare a static middleware to serve the built React files.
	app.Static("/", "./web/build")

	// API routes group
	groupApiV1 := app.Group("/api/v1", handlers.ApiLimiter)
	{
		groupApiV1.Get("/", handlers.Index())
	}
	// middlewares
	logQueue := make(chan models.Log, 5000)
	stop := make(chan struct{})
	go logs.StartLogWorker(s.LogDbConn, logQueue, 200, 5*time.Second, stop)
	app.Use(logs.LogMiddleware(s.LogDbConn, logQueue))

	groupApiV1.Get("/swagger/*", swagger.HandlerDefault)
	routerResource := handlers.NewRouterResources(s.JwtResources.JwtKeyfunc, s.MainDbConn, s.RedisStorage)
	// App Repository
	userRepository := user.NewUserRepository(s.MainDbConn)
	authRepository := auth.NewAuthRepository(s.MainDbConn, s.JwtResources, s.RedisStorage)
	cameraRepository := camera.NewCameraRepository(s.MainDbConn)
	detectRepository := detect.NewDetectRepository(s.MainDbConn)
	attackRepository := attack.NewAttackRepository(s.MainDbConn)

	// auto migrate DB only on main process
	if !fiber.IsChild() {
		if migrateErr := s.AutoMigrate(); migrateErr != nil {
			fmt.Printf("error while migrate book DB:\n %+v", migrateErr)
		}
	}

	// App Services
	userService := user.NewUserService(userRepository)
	authService := auth.NewAuthService(authRepository, userRepository)
	cameraService := camera.NewCameraService(cameraRepository)
	detectService := detect.NewDetectService(detectRepository)
	attackService := attack.NewAttackService(attackRepository)

	// MQTT Service for sending commands to Raspberry PI
	var mqttService *mqtt.Service
	if s.MQTTClient != nil && s.MQTTClient.IsConnected() {
		// Use separate topic for commands (topgun/command)
		mqttCommandTopic := viper.GetString("mqtt.command_topic")
		if mqttCommandTopic == "" {
			mqttCommandTopic = "topgun/command"
		}
		mqttService = mqtt.NewService(s.MQTTClient, mqttCommandTopic)
		// No need to subscribe to command topic (we only publish)
		log.Printf("MQTT command service initialized on topic: %s", mqttCommandTopic)

		// Start MQTT detection subscription for RaspberryPI data
		// Get default camera ID from config or create one
		detectMQTTTopic := viper.GetString("mqtt.detect_topic")
		if detectMQTTTopic == "" {
			detectMQTTTopic = "topgun/ai"
		}
		mqttBroker := viper.GetString("mqtt.broker")
		if mqttBroker == "" {
			mqttBroker = "tcp://localhost:1883"
		}

		// Get or create default camera for MQTT detections
		defaultCameraID := viper.GetString("mqtt.camera_id")
		var cameraUUID uuid.UUID
		if defaultCameraID != "" {
			if parsed, err := uuid.Parse(defaultCameraID); err == nil {
				cameraUUID = parsed
			}
		}
		// If no camera ID in config, use a fixed UUID for RaspberryPI camera
		if cameraUUID == uuid.Nil {
			// Fixed UUID for RaspberryPI MQTT camera
			cameraUUID = uuid.MustParse("3a939700-7724-4dc8-a5d8-47130aa68213")
		}

		// Start MQTT subscription in background
		go func() {
			if err := detect.StartMQTTSubscription(mqttBroker, detectMQTTTopic, cameraUUID, detectService); err != nil {
				fmt.Printf("Warning: failed to start MQTT detection subscription: %v\n", err)
			}
		}()
	}

	// App Routes
	auth.NewAuthHandler(groupApiV1.Group("/auth"), routerResource, authService, userService)
	user.NewUserHandler(groupApiV1.Group("/users"), routerResource, userService, authService)
	camera.NewCameraHandler(groupApiV1.Group("/camera"), routerResource, cameraService)
	detect.NewDetectHandler(groupApiV1.Group("/detect"), detectService)
	attack.NewAttackHandler(groupApiV1.Group("/attack"), attackService)

	// WebSocket routes for video streaming
	videoHandler := detect.NewDetectHandlerForWebSocket()
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws/video-input", videoHandler.HandleVideoInput())   // Python sends video here
	app.Get("/ws/video-stream", videoHandler.HandleVideoStream()) // Clients view video here

	// MQTT Routes
	if mqttService != nil {
		mqttHandler := mqtt.NewHandler(mqttService)
		mqtt.SetupRoutes(groupApiV1.Group("/mqtt"), mqttHandler)
	}

	// Log routes
	groupApiV1.Get("/logs", logs.GetLogsHandler(s.LogDbConn))
	groupApiV1.Get("/logs/stats", logs.GetLogStatsHandler(s.LogDbConn))

	// Serve log viewer HTML page
	groupApiV1.Get("/logs/viewer", func(c *fiber.Ctx) error {
		return c.SendFile("./web/logs.html")
	})

	// Prepare a fallback route to always serve the 'index.html', had there not be any matching routes.
	app.Static("*", "./web/build/index.html")
}
