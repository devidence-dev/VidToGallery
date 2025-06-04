package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/sirupsen/logrus"

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

	// API routes
	api := app.Group("/api/v1")
	
	api.Post("/process", handler.ProcessVideo)
	api.Get("/health", handler.HealthCheck)
}
