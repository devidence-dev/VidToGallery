package api

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"

	"vidtogallery/internal/models"
	"vidtogallery/pkg/downloader"
)

type Handler struct {
	downloaderService *downloader.Service
	logger            *logrus.Logger
}

func NewHandler(downloaderService *downloader.Service, logger *logrus.Logger) *Handler {
	return &Handler{
		downloaderService: downloaderService,
		logger:            logger,
	}
}

// ProcessVideo processes a video URL and returns download information
// @Summary Process video URL
// @Description Extract video download URL from social media platform
// @Tags Video Processing
// @Accept json
// @Produce json
// @Param request body models.VideoRequest true "Video URL to process"
// @Success 200 {object} models.VideoResponse "Video processed successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /api/v1/process [post]
func (h *Handler) ProcessVideo(c *fiber.Ctx) error {
	var req models.VideoRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WithError(err).Error("Failed to parse request body")
		return c.Status(400).JSON(models.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_REQUEST",
		})
	}

	if req.URL == "" {
		return c.Status(400).JSON(models.ErrorResponse{
			Error: "URL is required",
			Code:  "MISSING_URL",
		})
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()

	h.logger.WithField("url", req.URL).Info("Processing video URL")

	response, err := h.downloaderService.ProcessURL(ctx, req.URL)
	if err != nil {
		h.logger.WithError(err).WithField("url", req.URL).Error("Failed to process video")
		return c.Status(500).JSON(models.ErrorResponse{
			Error:   "Failed to process video",
			Code:    "PROCESSING_ERROR",
			Details: err.Error(),
		})
	}

	h.logger.WithFields(logrus.Fields{
		"url":       req.URL,
		"platform":  response.Platform,
		"video_url": response.VideoURL,
	}).Info("Video processed successfully")

	return c.JSON(response)
}

// HealthCheck returns the health status of the API
// @Summary Health check endpoint
// @Description Check if the API is running and healthy
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]interface{} "API is healthy"
// @Router /health [get]
func (h *Handler) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "ok",
		"timestamp": time.Now(),
		"service":   "vidtogallery",
	})
}
