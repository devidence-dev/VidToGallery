package downloader

import (
	"context"
	"fmt"
	"sync"

	"vidtogallery/internal/models"
	"vidtogallery/pkg/cache"
	"vidtogallery/pkg/config"
)

type Downloader interface {
	ValidateURL(url string) bool
	ExtractVideoURL(url string) (*models.VideoResponse, error)
}

type Service struct {
	downloaders  map[string]Downloader
	workers      chan struct{}
	mu           sync.RWMutex
	cacheService *cache.Service
}

func NewService(maxConcurrent int, cfg *config.Config, cacheService *cache.Service) *Service {
	service := &Service{
		downloaders:  make(map[string]Downloader),
		workers:      make(chan struct{}, maxConcurrent),
		cacheService: cacheService,
	}

	// Register downloaders with configuration
	service.RegisterDownloader("instagram", NewInstagramDownloaderWithConfig(cfg))
	service.RegisterDownloader("twitter", NewTwitterDownloaderWithConfig(cfg))

	return service
}

func (s *Service) RegisterDownloader(platform string, downloader Downloader) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.downloaders[platform] = downloader
}

func (s *Service) DetectPlatform(url string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for platform, downloader := range s.downloaders {
		if downloader.ValidateURL(url) {
			return platform
		}
	}
	return ""
}

func (s *Service) ProcessURL(ctx context.Context, url string) (*models.VideoResponse, error) {
	// Try to get from cache first
	if cachedVideo, found := s.cacheService.GetVideo(ctx, url); found {
		return cachedVideo, nil
	}

	// Acquire worker slot
	select {
	case s.workers <- struct{}{}:
		defer func() { <-s.workers }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	platform := s.DetectPlatform(url)
	if platform == "" {
		return nil, fmt.Errorf("unsupported platform or invalid URL")
	}

	s.mu.RLock()
	downloader := s.downloaders[platform]
	s.mu.RUnlock()

	video, err := downloader.ExtractVideoURL(url)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := s.cacheService.SetVideo(ctx, url, video); err != nil {
		// Log error but don't fail the request
	}

	return video, nil
}
