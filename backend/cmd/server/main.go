package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"

	"vidtogallery/pkg/api"
	"vidtogallery/pkg/config"
	"vidtogallery/pkg/downloader"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Setup logger
	logger := logrus.New()
	if cfg.Environment == "development" {
		logger.SetLevel(logrus.DebugLevel)
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	} else {
		logger.SetLevel(logrus.InfoLevel)
		logger.SetFormatter(&logrus.JSONFormatter{})
	}

	// Initialize downloader service
	downloaderService := downloader.NewService(cfg.Download.MaxConcurrent, cfg)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "VidToGallery API",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			logger.WithError(err).Error("Request error")
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
				"code":  code,
			})
		},
	})

	// Setup routes
	api.SetupRoutes(app, downloaderService, logger)

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	logger.WithField("address", addr).Info("Starting server")

	if err := app.Listen(addr); err != nil {
		logger.WithError(err).Fatal("Failed to start server")
	}
}
