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

type YouTubeDownloader struct {
	client    *http.Client
	uaRotator *useragent.Rotator
}

var youtubeRegex = regexp.MustCompile(`^(?:https?://)?(?:www\.)?(youtube\.com/watch\?v=|youtu\.be/|youtube\.com/shorts/)([a-zA-Z0-9_-]{11})`)

func NewYouTubeDownloader() *YouTubeDownloader {
	return &YouTubeDownloader{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		uaRotator: useragent.NewRotator(true),
	}
}

func NewYouTubeDownloaderWithConfig(cfg *config.Config) *YouTubeDownloader {
	return &YouTubeDownloader{
		client: &http.Client{
			Timeout: cfg.Download.Timeout,
		},
		uaRotator: useragent.NewRotator(cfg.UserAgent.RandomOrder),
	}
}

func (d *YouTubeDownloader) ValidateURL(url string) bool {
	return youtubeRegex.MatchString(url)
}

func (d *YouTubeDownloader) ExtractVideoURL(url string) (*models.VideoResponse, error) {
	if !d.ValidateURL(url) {
		return nil, fmt.Errorf("invalid YouTube URL")
	}

	// Get next user agent
	userAgent := d.uaRotator.Next()

	// Extract video ID from URL
	matches := youtubeRegex.FindStringSubmatch(url)
	if len(matches) < 3 {
		return nil, fmt.Errorf("could not extract video ID from URL")
	}
	videoID := matches[2]

	// Get video page to extract player config
	videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
	req, err := http.NewRequest("GET", videoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch YouTube page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("YouTube returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	videoStreamURL, title, err := d.parseYouTubeResponse(string(body), videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YouTube response: %w", err)
	}

	return &models.VideoResponse{
		VideoURL:    videoStreamURL,
		Title:       title,
		Platform:    "youtube",
		Quality:     "auto",
		ProcessedAt: time.Now(),
		Metadata: map[string]string{
			"source":   url,
			"video_id": videoID,
		},
	}, nil
}

func (d *YouTubeDownloader) parseYouTubeResponse(html string, videoID string) (string, string, error) {
	// Extract title from page title tag
	titleStart := strings.Index(html, "<title>")
	titleEnd := strings.Index(html, "</title>")
	var title string
	if titleStart != -1 && titleEnd != -1 {
		title = html[titleStart+7 : titleEnd]
		title = strings.TrimSuffix(title, " - YouTube")
	}

	// Look for ytInitialPlayerResponse
	playerResponseStart := strings.Index(html, "var ytInitialPlayerResponse = ")
	if playerResponseStart == -1 {
		playerResponseStart = strings.Index(html, "ytInitialPlayerResponse\":")
		if playerResponseStart == -1 {
			return "", "", fmt.Errorf("could not find player response in page")
		}
		playerResponseStart += len("ytInitialPlayerResponse\":")
	} else {
		playerResponseStart += len("var ytInitialPlayerResponse = ")
	}

	// Find the end of the JSON object
	playerResponseEnd := strings.Index(html[playerResponseStart:], ";</script>")
	if playerResponseEnd == -1 {
		playerResponseEnd = strings.Index(html[playerResponseStart:], ";var ")
		if playerResponseEnd == -1 {
			playerResponseEnd = strings.Index(html[playerResponseStart:], ",\"")
			if playerResponseEnd == -1 {
				return "", "", fmt.Errorf("could not find end of player response")
			}
		}
	}

	playerResponseJSON := html[playerResponseStart : playerResponseStart+playerResponseEnd]

	// Parse the JSON
	var playerResponse map[string]interface{}
	if err := json.Unmarshal([]byte(playerResponseJSON), &playerResponse); err != nil {
		return "", "", fmt.Errorf("failed to parse player response JSON: %w", err)
	}

	// Navigate through the JSON structure to find streaming data
	streamingData, ok := playerResponse["streamingData"].(map[string]interface{})
	if !ok {
		return "", "", fmt.Errorf("no streaming data found")
	}

	// Try adaptive formats first (better quality)
	if adaptiveFormats, ok := streamingData["adaptiveFormats"].([]interface{}); ok {
		for _, format := range adaptiveFormats {
			formatMap, ok := format.(map[string]interface{})
			if !ok {
				continue
			}

			mimeType, ok := formatMap["mimeType"].(string)
			if !ok || !strings.Contains(mimeType, "video/mp4") {
				continue
			}

			if url, ok := formatMap["url"].(string); ok {
				return url, title, nil
			}
		}
	}

	// Fallback to regular formats
	if formats, ok := streamingData["formats"].([]interface{}); ok {
		for _, format := range formats {
			formatMap, ok := format.(map[string]interface{})
			if !ok {
				continue
			}

			mimeType, ok := formatMap["mimeType"].(string)
			if !ok || !strings.Contains(mimeType, "video/mp4") {
				continue
			}

			if url, ok := formatMap["url"].(string); ok {
				return url, title, nil
			}
		}
	}

	return "", "", fmt.Errorf("no suitable video format found")
}
