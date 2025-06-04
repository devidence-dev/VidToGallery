package downloader

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"vidtogallery/internal/models"
	"vidtogallery/pkg/config"
	"vidtogallery/pkg/useragent"
)

type InstagramDownloader struct {
	client    *http.Client
	uaRotator *useragent.Rotator
}

var instagramRegex = regexp.MustCompile(`^(?:https?://)?(?:www\.)?instagram\.com/(?:p|reel)/([A-Za-z0-9_-]+)/?`)

func NewInstagramDownloader() *InstagramDownloader {
	return &InstagramDownloader{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		uaRotator: useragent.NewRotator(true), // Use random rotation
	}
}

func NewInstagramDownloaderWithConfig(cfg *config.Config) *InstagramDownloader {
	return &InstagramDownloader{
		client: &http.Client{
			Timeout: cfg.Download.Timeout,
		},
		uaRotator: useragent.NewRotator(cfg.UserAgent.RandomOrder),
	}
}

func (d *InstagramDownloader) ValidateURL(url string) bool {
	return instagramRegex.MatchString(url)
}

func (d *InstagramDownloader) ExtractVideoURL(url string) (*models.VideoResponse, error) {
	if !d.ValidateURL(url) {
		return nil, fmt.Errorf("invalid Instagram URL")
	}

	// Get next user agent
	userAgent := d.uaRotator.Next()

	// Add ?__a=1 to get JSON response
	apiURL := strings.TrimSuffix(url, "/") + "/?__a=1&__d=dis"

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Instagram data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Instagram returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	videoURL, title, err := d.parseInstagramResponse(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Instagram response: %w", err)
	}

	return &models.VideoResponse{
		VideoURL:    videoURL,
		Title:       title,
		Platform:    "instagram",
		Quality:     "auto",
		ProcessedAt: time.Now(),
		Metadata: map[string]string{
			"source": url,
		},
	}, nil
}

func (d *InstagramDownloader) parseInstagramResponse(data []byte) (string, string, error) {
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Navigate through Instagram's JSON structure to find video URL
	items, ok := result["items"].([]interface{})
	if !ok || len(items) == 0 {
		return "", "", fmt.Errorf("no items found in response")
	}

	item := items[0].(map[string]interface{})

	// Try to get video URL
	var videoURL string
	var title string

	// Check if it's a video post
	if videoVersions, ok := item["video_versions"].([]interface{}); ok && len(videoVersions) > 0 {
		// Get the highest quality video
		video := videoVersions[0].(map[string]interface{})
		videoURL = video["url"].(string)
	} else {
		return "", "", fmt.Errorf("no video found in this post")
	}

	// Get title/caption
	if caption := item["caption"]; caption != nil {
		if captionMap, ok := caption.(map[string]interface{}); ok {
			if text, ok := captionMap["text"].(string); ok {
				title = text
			}
		}
	}

	if videoURL == "" {
		return "", "", fmt.Errorf("video URL not found")
	}

	return videoURL, title, nil
}
