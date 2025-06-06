package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"vidtogallery/internal/models"
	"vidtogallery/pkg/config"
)

var (
	ErrCacheDisabled = errors.New("cache is disabled")
	ErrCacheNotFound = errors.New("key not found in cache")
)

type Service struct {
	client   *redis.Client
	logger   *logrus.Logger
	videoTTL time.Duration
}

func NewService(cfg *config.Config, logger *logrus.Logger) *Service {
	opts, err := redis.ParseURL(cfg.Redis.URL)
	if err != nil {
		logger.WithError(err).Fatal("Failed to parse Redis URL")
	}

	if cfg.Redis.Password != "" {
		opts.Password = cfg.Redis.Password
	}
	opts.DB = cfg.Redis.DB

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger.WithError(err).Warn("Redis connection failed, caching will be disabled")
		return &Service{
			client:   nil,
			logger:   logger,
			videoTTL: cfg.Cache.VideoTTL,
		}
	}

	logger.Info("Redis connection established")
	return &Service{
		client:   client,
		logger:   logger,
		videoTTL: cfg.Cache.VideoTTL,
	}
}

func (s *Service) GetVideo(ctx context.Context, url string) (*models.VideoResponse, bool) {
	if s.client == nil {
		return nil, false
	}

	key := s.videoKey(url)
	data, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err != redis.Nil {
			s.logger.WithError(err).WithField("key", key).Error("Failed to get from cache")
		}
		return nil, false
	}

	var video models.VideoResponse
	if err := json.Unmarshal([]byte(data), &video); err != nil {
		s.logger.WithError(err).WithField("key", key).Error("Failed to unmarshal cached video")
		return nil, false
	}

	s.logger.WithField("url", url).Debug("Video found in cache")
	return &video, true
}

func (s *Service) SetVideo(ctx context.Context, url string, video *models.VideoResponse) error {
	if s.client == nil {
		return nil
	}

	key := s.videoKey(url)
	data, err := json.Marshal(video)
	if err != nil {
		return err
	}

	if err := s.client.Set(ctx, key, data, s.videoTTL).Err(); err != nil {
		s.logger.WithError(err).WithField("key", key).Error("Failed to set cache")
		return err
	}

	s.logger.WithFields(logrus.Fields{
		"url": url,
		"ttl": s.videoTTL,
		"key": key,
	}).Debug("Video cached successfully")

	return nil
}

// GetVideoFile retrieves a cached video file
func (s *Service) GetVideoFile(ctx context.Context, videoURL string) ([]byte, error) {
	if s.client == nil {
		return nil, ErrCacheDisabled
	}

	key := s.videoFileKey(videoURL)
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrCacheNotFound
		}
		return nil, err
	}

	return data, nil
}

// CacheVideoFile stores a video file in cache with custom TTL
func (s *Service) CacheVideoFile(ctx context.Context, videoURL string, data []byte, ttl time.Duration) error {
	if s.client == nil {
		return ErrCacheDisabled
	}

	key := s.videoFileKey(videoURL)
	return s.client.Set(ctx, key, data, ttl).Err()
}

func (s *Service) videoKey(url string) string {
	return "video:" + url
}

// videoFileKey generates a cache key for video files
func (s *Service) videoFileKey(videoURL string) string {
	return "video_file:" + videoURL
}

func (s *Service) Close() error {
	if s.client == nil {
		return nil
	}
	return s.client.Close()
}
