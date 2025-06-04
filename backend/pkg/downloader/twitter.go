package downloader

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"vidtogallery/internal/models"
	"vidtogallery/pkg/config"
	"vidtogallery/pkg/useragent"
)

type TwitterDownloader struct {
	client    *http.Client
	uaRotator *useragent.Rotator
}

var twitterRegex = regexp.MustCompile(`^(?:https?://)?(?:www\.)?(twitter\.com|x\.com)/\w+/status/(\d+)`)

func NewTwitterDownloader() *TwitterDownloader {
	return &TwitterDownloader{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		uaRotator: useragent.NewRotator(true),
	}
}

func NewTwitterDownloaderWithConfig(cfg *config.Config) *TwitterDownloader {
	return &TwitterDownloader{
		client: &http.Client{
			Timeout: cfg.Download.Timeout,
		},
		uaRotator: useragent.NewRotator(cfg.UserAgent.RandomOrder),
	}
}

func (d *TwitterDownloader) ValidateURL(url string) bool {
	return twitterRegex.MatchString(url)
}

func (d *TwitterDownloader) ExtractVideoURL(url string) (*models.VideoResponse, error) {
	if !d.ValidateURL(url) {
		return nil, fmt.Errorf("invalid Twitter/X URL")
	}

	// Get next user agent
	userAgent := d.uaRotator.Next()

	// Extract tweet ID from URL
	matches := twitterRegex.FindStringSubmatch(url)
	if len(matches) < 3 {
		return nil, fmt.Errorf("could not extract tweet ID from URL")
	}
	tweetID := matches[2]

	// Use Twitter's API endpoint for guest access
	apiURL := fmt.Sprintf("https://api.twitter.com/1.1/statuses/show.json?id=%s&include_entities=true", tweetID)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Twitter data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Twitter API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	videoURL, title, err := d.parseTwitterResponse(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Twitter response: %w", err)
	}

	return &models.VideoResponse{
		VideoURL:    videoURL,
		Title:       title,
		Platform:    "twitter",
		Quality:     "auto",
		ProcessedAt: time.Now(),
		Metadata: map[string]string{
			"source":   url,
			"tweet_id": tweetID,
		},
	}, nil
}

func (d *TwitterDownloader) parseTwitterResponse(data []byte) (string, string, error) {
	var tweet map[string]interface{}
	if err := json.Unmarshal(data, &tweet); err != nil {
		return "", "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Get tweet text
	text, _ := tweet["text"].(string)

	// Look for video in extended_entities
	extendedEntities, ok := tweet["extended_entities"].(map[string]interface{})
	if !ok {
		return "", "", fmt.Errorf("no media entities found in tweet")
	}

	media, ok := extendedEntities["media"].([]interface{})
	if !ok || len(media) == 0 {
		return "", "", fmt.Errorf("no media found in tweet")
	}

	// Find video media
	for _, mediaItem := range media {
		mediaMap, ok := mediaItem.(map[string]interface{})
		if !ok {
			continue
		}

		mediaType, ok := mediaMap["type"].(string)
		if !ok || (mediaType != "video" && mediaType != "animated_gif") {
			continue
		}

		// Get video info
		videoInfo, ok := mediaMap["video_info"].(map[string]interface{})
		if !ok {
			continue
		}

		variants, ok := videoInfo["variants"].([]interface{})
		if !ok || len(variants) == 0 {
			continue
		}

		// Find the best quality video URL
		var bestURL string
		var bestBitrate int

		for _, variant := range variants {
			variantMap, ok := variant.(map[string]interface{})
			if !ok {
				continue
			}

			contentType, ok := variantMap["content_type"].(string)
			if !ok || contentType != "video/mp4" {
				continue
			}

			url, ok := variantMap["url"].(string)
			if !ok {
				continue
			}

			bitrate, _ := variantMap["bitrate"].(float64)
			if bestURL == "" || int(bitrate) > bestBitrate {
				bestURL = url
				bestBitrate = int(bitrate)
			}
		}

		if bestURL != "" {
			return bestURL, text, nil
		}
	}

	return "", "", fmt.Errorf("no video URL found in tweet")
}
