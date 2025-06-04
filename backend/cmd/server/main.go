// VidToGallery API
//
// API for downloading videos from social media platforms (Instagram, Twitter/X, YouTube)
//
//	Title: VidToGallery API
//	Description: Extract video URLs from social media platforms for direct download
//	Version: 1.0.0
//	Host: localhost:8080
//	BasePath: /
//	Schemes: http, https
//
//	Contact: VidToGallery API Support <support@vidtogallery.com>
//
// @title VidToGallery API
// @description API for downloading videos from social media platforms (Instagram, Twitter/X, YouTube)
// @version 1.0.0
// @host localhost:8080
// @BasePath /
// @schemes http https
// @contact.name VidToGallery API Support
// @contact.email support@vidtogallery.com
package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"

	_ "vidtogallery/docs" // Import generated docs
	"vidtogallery/pkg/api"
	"vidtogallery/pkg/cache"
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

	// Initialize cache service
	cacheService := cache.NewService(cfg, logger)
	defer cacheService.Close()

	// Initialize downloader service
	downloaderService := downloader.NewService(cfg.Download.MaxConcurrent, cfg, cacheService)

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
