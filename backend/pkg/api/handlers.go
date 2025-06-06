package api

import (
	"context"
	"fmt"
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

// DownloadVideo downloads a video with specified quality
// @Summary Download video with quality
// @Description Download video from social media platform with specified quality
// @Tags Video Processing
// @Accept json
// @Produce json
// @Param request body models.VideoRequest true "Video URL and quality to download"
// @Success 200 {object} models.VideoResponse "Video downloaded successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /api/v1/download [post]
func (h *Handler) DownloadVideo(c *fiber.Ctx) error {
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

	h.logger.WithField("url", req.URL).Info("Downloading video")

	// Set default quality to "best" if not specified
	quality := req.Quality
	if quality == "" {
		quality = "best"
	}

	// Download video with specified quality
	response, err := h.downloaderService.ProcessURLWithQuality(ctx, req.URL, quality)
	if err != nil {
		h.logger.WithError(err).WithField("url", req.URL).Error("Failed to download video")
		return c.Status(500).JSON(models.ErrorResponse{
			Error:   "Failed to download video",
			Code:    "DOWNLOAD_ERROR",
			Details: err.Error(),
		})
	}

	h.logger.WithFields(logrus.Fields{
		"url":       req.URL,
		"platform":  response.Platform,
		"video_url": response.VideoURL,
		"quality":   quality,
	}).Info("Video downloaded successfully")

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

// GetQualities returns available video qualities for a given URL
// @Summary Get available video qualities
// @Description Get list of available video qualities for a social media URL
// @Tags Video Processing
// @Accept json
// @Produce json
// @Param request body models.QualityRequest true "Video URL to check qualities for"
// @Success 200 {object} models.QualitiesResponse "Available qualities retrieved successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /api/v1/qualities [post]
func (h *Handler) GetQualities(c *fiber.Ctx) error {
	var req models.QualityRequest
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

	h.logger.WithField("url", req.URL).Info("Getting available qualities")

	// Get available qualities for the video
	response, err := h.downloaderService.GetAvailableQualities(ctx, req.URL)
	if err != nil {
		h.logger.WithError(err).WithField("url", req.URL).Error("Failed to get available qualities")
		return c.Status(500).JSON(models.ErrorResponse{
			Error:   "Failed to get available qualities",
			Code:    "QUALITIES_ERROR",
			Details: err.Error(),
		})
	}

	h.logger.WithFields(logrus.Fields{
		"url":       req.URL,
		"platform":  response.Platform,
		"qualities": len(response.AvailableQualities),
	}).Info("Qualities retrieved successfully")

	return c.JSON(response)
}

// ProxyDownload downloads video file through backend to avoid CORS issues
// @Summary Proxy download video file
// @Description Download video file through backend proxy to avoid CORS restrictions
// @Tags Video Processing
// @Accept json
// @Produce application/octet-stream
// @Param request body models.ProxyDownloadRequest true "Video URL to proxy download"
// @Success 200 {file} binary "Video file"
// @Failure 400 {object} models.ErrorResponse "Invalid request"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /api/v1/proxy-download [post]
func (h *Handler) ProxyDownload(c *fiber.Ctx) error {
	var req models.ProxyDownloadRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WithError(err).Error("Failed to parse request body")
		return c.Status(400).JSON(models.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_REQUEST",
		})
	}

	if req.VideoURL == "" {
		return c.Status(400).JSON(models.ErrorResponse{
			Error: "video_url is required",
			Code:  "MISSING_VIDEO_URL",
		})
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Context(), 60*time.Second)
	defer cancel()

	h.logger.WithField("video_url", req.VideoURL).Info("Proxying video download")

	// Proxy download through downloader service
	response, err := h.downloaderService.ProxyDownload(ctx, req.VideoURL)
	if err != nil {
		h.logger.WithError(err).WithField("video_url", req.VideoURL).Error("Failed to proxy download video")
		return c.Status(500).JSON(models.ErrorResponse{
			Error:   "Failed to download video",
			Code:    "PROXY_DOWNLOAD_ERROR",
			Details: err.Error(),
		})
	}

	// Set appropriate headers
	c.Set("Content-Type", "application/octet-stream")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"video_%d.mp4\"", time.Now().Unix()))

	h.logger.WithField("video_url", req.VideoURL).Info("Video proxy download completed")

	return c.Send(response.Data)
}
