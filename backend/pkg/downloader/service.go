package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"vidtogallery/internal/models"
	"vidtogallery/pkg/cache"
	"vidtogallery/pkg/config"
)

type Downloader interface {
	ValidateURL(url string) bool
	ExtractVideoURL(url string) (*models.VideoResponse, error)
	ExtractVideoURLWithQuality(url string, quality string) (*models.VideoResponse, error)
}

type Service struct {
	downloader   *UniversalDownloader
	workers      chan struct{}
	mu           sync.RWMutex
	cacheService *cache.Service
}

func NewService(maxConcurrent int, cfg *config.Config, cacheService *cache.Service) *Service {
	return &Service{
		downloader:   NewUniversalDownloaderWithConfig(cfg),
		workers:      make(chan struct{}, maxConcurrent),
		cacheService: cacheService,
	}
}

func (s *Service) DetectPlatform(url string) string {
	return s.downloader.DetectPlatform(url)
}

func (s *Service) ProcessURL(ctx context.Context, url string) (*models.VideoResponse, error) {
	return s.ProcessURLWithQuality(ctx, url, "best")
}

func (s *Service) ProcessURLWithQuality(ctx context.Context, url string, quality string) (*models.VideoResponse, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("%s:%s", url, quality)
	if cachedVideo, found := s.cacheService.GetVideo(ctx, cacheKey); found {
		return cachedVideo, nil
	}

	// Acquire worker slot
	select {
	case s.workers <- struct{}{}:
		defer func() { <-s.workers }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Use universal downloader
	video, err := s.downloader.ExtractVideoURLWithQuality(url, quality)
	if err != nil {
		return nil, err
	}

	// Cache the result with quality-specific key
	if err := s.cacheService.SetVideo(ctx, cacheKey, video); err != nil {
		// Log error but don't fail the request
	}

	return video, nil
}

func (s *Service) GetAvailableQualities(ctx context.Context, url string) (*models.QualitiesResponse, error) {
	return s.downloader.GetAvailableQualities(url)
}

// ProxyDownload downloads video through backend to avoid CORS issues
func (s *Service) ProxyDownload(ctx context.Context, videoURL string) (*models.ProxyDownloadResponse, error) {
	s.workers <- struct{}{}
	defer func() { <-s.workers }()

	// Check cache first
	if s.cacheService != nil {
		if cachedData, err := s.cacheService.GetVideoFile(ctx, videoURL); err == nil && cachedData != nil {
			return &models.ProxyDownloadResponse{
				Data: cachedData,
			}, nil
		}
	}

	// Download video file
	response, err := http.Get(videoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download video: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download video: HTTP %d", response.StatusCode)
	}

	// Read video data
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read video data: %w", err)
	}

	// Cache the video file with 5 minute expiration
	if s.cacheService != nil {
		if err := s.cacheService.CacheVideoFile(ctx, videoURL, data, 5*time.Minute); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to cache video file: %v\n", err)
		}
	}

	return &models.ProxyDownloadResponse{
		Data: data,
	}, nil
}
