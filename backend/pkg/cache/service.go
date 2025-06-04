package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"vidtogallery/internal/models"
	"vidtogallery/pkg/config"
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

func (s *Service) videoKey(url string) string {
	return "video:" + url
}

func (s *Service) Close() error {
	if s.client == nil {
		return nil
	}
	return s.client.Close()
}
