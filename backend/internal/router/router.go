package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"monorepo/backend/internal/config"
)

// RouteRegistrar is implemented by every feature module.
type RouteRegistrar interface {
	RegisterRoutes(router fiber.Router)
}

func Setup(cfg *config.Config, modules ...RouteRegistrar) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "Monorepo Go Fiber API",
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSOrigins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, PATCH, DELETE, OPTIONS",
	}))

	api := app.Group("/api/v1")
	for _, mod := range modules {
		mod.RegisterRoutes(api)
	}

	return app
}
