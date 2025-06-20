package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/sirupsen/logrus"
	fiberSwagger "github.com/swaggo/fiber-swagger"

	"vidtogallery/pkg/downloader"
)

func SetupRoutes(app *fiber.App, downloaderService *downloader.Service, log *logrus.Logger) {
	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept",
	}))

	// Initialize handlers
	handler := NewHandler(downloaderService, log)

	// Swagger documentation
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// Health endpoint (outside of API group for simple access)
	app.Get("/health", handler.HealthCheck)

	// API routes
	api := app.Group("/api/v1")
	api.Post("/download", handler.DownloadVideo)
	api.Post("/qualities", handler.GetQualities)
	api.Post("/proxy-download", handler.ProxyDownload)
}
