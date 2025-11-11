package infrastructure

import (
	"fmt"
	"time"

	"topgun-services/internal/handlers"
	"topgun-services/pkg/auth"
	"topgun-services/pkg/camera"
	"topgun-services/pkg/detect"
	"topgun-services/pkg/logs"
	"topgun-services/pkg/models"
	"topgun-services/pkg/user"

	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/fiber/v2"
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
	// App Routes
	auth.NewAuthHandler(groupApiV1.Group("/auth"), routerResource, authService, userService)
	user.NewUserHandler(groupApiV1.Group("/users"), routerResource, userService, authService)
	camera.NewCameraHandler(groupApiV1.Group("/camera"), routerResource, cameraService)
	detect.NewDetectHandler(groupApiV1.Group("/detect"), detectService)

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
