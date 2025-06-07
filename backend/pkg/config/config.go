package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server struct {
		Port string
		Host string
	}
	Redis struct {
		URL      string
		Password string
		DB       int
	}
	Cache struct {
		TTL      time.Duration
		VideoTTL time.Duration
	}
	Download struct {
		MaxConcurrent int
		Timeout       time.Duration
	}
	UserAgent struct {
		RotateAgents bool
		RandomOrder  bool
	}
	Environment string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		// Don't return error if .env file doesn't exist
	}

	cfg := &Config{}

	cfg.Server.Port = getEnv("PORT", "8080")
	cfg.Server.Host = getEnv("HOST", "0.0.0.0")

	cfg.Redis.URL = getEnv("REDIS_URL", "redis://localhost:6379")
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "")
	cfg.Redis.DB = getEnvAsInt("REDIS_DB", 0)

	cfg.Cache.TTL = getEnvAsDuration("CACHE_TTL", 24*time.Hour)
	cfg.Cache.VideoTTL = getEnvAsDuration("VIDEO_CACHE_TTL", 24*time.Hour)

	cfg.Download.MaxConcurrent = getEnvAsInt("MAX_CONCURRENT_DOWNLOADS", 5)
	cfg.Download.Timeout = getEnvAsDuration("DOWNLOAD_TIMEOUT", 30*time.Second)

	cfg.UserAgent.RotateAgents = getEnvAsBool("ROTATE_USER_AGENTS", true)
	cfg.UserAgent.RandomOrder = getEnvAsBool("RANDOM_USER_AGENT_ORDER", true)

	cfg.Environment = getEnv("ENV", "development")

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
