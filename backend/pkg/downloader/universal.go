package downloader

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"vidtogallery/internal/models"
	"vidtogallery/pkg/config"
	"vidtogallery/pkg/quality"
	"vidtogallery/pkg/useragent"
)

// UniversalDownloader handles video extraction from any platform supported by yt-dlp
type UniversalDownloader struct {
	uaRotator      *useragent.Rotator
	qualityManager *quality.Manager
	ytdlpPath      string
}

// UniversalYtDlpInfo represents the JSON structure returned by yt-dlp
type UniversalYtDlpInfo struct {
	URL         string                 `json:"url"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Duration    float64                `json:"duration"`
	Thumbnail   string                 `json:"thumbnail"`
	Formats     []UniversalYtDlpFormat `json:"formats,omitempty"`
}

type UniversalYtDlpFormat struct {
	FormatID string `json:"format_id"`
	URL      string `json:"url"`
	Height   int    `json:"height"`
	Width    int    `json:"width"`
	VCodec   string `json:"vcodec"`
	ACodec   string `json:"acodec"`
	FileSize int64  `json:"filesize,omitempty"`
}

// Supported platforms with their regex patterns
var platformPatterns = map[string]*regexp.Regexp{
	"instagram": regexp.MustCompile(`^(?:https?://)?(?:www\.)?instagram\.com/(?:p|reel)/([A-Za-z0-9_-]+)/?`),
	"twitter":   regexp.MustCompile(`^(?:https?://)?(?:www\.)?(?:twitter\.com|x\.com)/[^/]+/status/(\d+)`),
	"tiktok":    regexp.MustCompile(`^(?:https?://)?(?:www\.)?(?:tiktok\.com|vm\.tiktok\.com)`),
}

// YouTube pattern for detecting and rejecting YouTube URLs
var youtubePattern = regexp.MustCompile(`^(?:https?://)?(?:www\.)?(?:youtube\.com/watch\?v=|youtu\.be/|youtube\.com/shorts/)([A-Za-z0-9_-]+)`)

func NewUniversalDownloader() *UniversalDownloader {
	return &UniversalDownloader{
		uaRotator:      useragent.NewRotator(true),
		qualityManager: quality.NewManager(),
		ytdlpPath:      "yt-dlp",
	}
}

func NewUniversalDownloaderWithConfig(cfg *config.Config) *UniversalDownloader {
	return &UniversalDownloader{
		uaRotator:      useragent.NewRotator(cfg.UserAgent.RandomOrder),
		qualityManager: quality.NewManager(),
		ytdlpPath:      "yt-dlp",
	}
}

func (d *UniversalDownloader) ValidateURL(url string) bool {
	// Clean the URL by trimming whitespace
	url = strings.TrimSpace(url)

	// Check if it's a YouTube URL and reject it
	if youtubePattern.MatchString(url) {
		return false
	}

	for _, pattern := range platformPatterns {
		if pattern.MatchString(url) {
			return true
		}
	}
	return false
}

func (d *UniversalDownloader) DetectPlatform(url string) string {
	// Clean the URL by trimming whitespace
	url = strings.TrimSpace(url)

	for platform, pattern := range platformPatterns {
		if pattern.MatchString(url) {
			return platform
		}
	}
	return "unknown"
}

// ExtractVideoURL extracts video URL with default "best" quality
func (d *UniversalDownloader) ExtractVideoURL(url string) (*models.VideoResponse, error) {
	return d.ExtractVideoURLWithQuality(url, "best")
}

// ExtractVideoURLWithQuality extracts video URL using yt-dlp
func (d *UniversalDownloader) ExtractVideoURLWithQuality(url string, quality string) (*models.VideoResponse, error) {
	// Clean the URL by trimming whitespace
	url = strings.TrimSpace(url)

	// Check if it's a YouTube URL and return specific error
	if youtubePattern.MatchString(url) {
		return nil, fmt.Errorf("YouTube videos are not supported at this time")
	}

	if !d.ValidateURL(url) {
		return nil, fmt.Errorf("unsupported URL or platform")
	}

	// Check if yt-dlp is available
	if _, err := exec.LookPath(d.ytdlpPath); err != nil {
		return nil, fmt.Errorf("yt-dlp not found in PATH: %w", err)
	}

	// Prepare yt-dlp command
	args := []string{
		"--no-check-certificate",
		"--no-warnings",
		"--dump-json",
		"--no-playlist",
	}

	// Add quality selection based on preference
	switch quality {
	case "best":
		args = append(args, "--format", "bestvideo[ext=mp4]+bestaudio/best")
	case "worst":
		args = append(args, "--format", "worstvideo[ext=mp4]+worstaudio/worst")
	default:
		// For specific quality formats like "best[height<=1080]", use proper yt-dlp syntax
		if quality != "" && strings.Contains(quality, "height<=") {
			// Extract height value and use proper format selection
			args = append(args, "--format", "bestvideo"+quality[4:]+"+bestaudio/best")
		} else if quality != "" {
			args = append(args, "--format", quality)
		} else {
			// Default to best quality if no quality specified
			args = append(args, "--format", "bestvideo[ext=mp4]+bestaudio/best")
		}
	}

	args = append(args, url)

	// Execute yt-dlp
	cmd := exec.Command(d.ytdlpPath, args...)

	// Log the command being executed for debugging
	fmt.Printf("DEBUG: Executing yt-dlp with args: %v\n", args)

	output, err := cmd.Output()
	if err != nil {
		// Get stderr for better error reporting
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			fmt.Printf("DEBUG: yt-dlp stderr: %s\n", stderr)
			return nil, fmt.Errorf("yt-dlp failed: %w, stderr: %s", err, stderr)
		}
		return nil, fmt.Errorf("yt-dlp failed: %w", err)
	}

	// Parse JSON output
	var info UniversalYtDlpInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, fmt.Errorf("failed to parse yt-dlp output: %w", err)
	}

	// Get the video URL - always check formats first for quality selection
	videoURL := ""

	// Try to find the best format matching the requested quality
	if len(info.Formats) > 0 {
		var selectedFormat *UniversalYtDlpFormat

		// Parse the quality parameter to understand what we're looking for
		switch {
		case quality == "best":
			// Find format with highest resolution
			maxHeight := 0
			for _, format := range info.Formats {
				if format.URL != "" && format.VCodec != "none" {
					if format.Height > maxHeight {
						maxHeight = format.Height
						selectedFormat = &format
					} else if selectedFormat == nil {
						selectedFormat = &format
					}
				}
			}
		case quality == "worst":
			// Find format with lowest resolution
			minHeight := 99999
			for _, format := range info.Formats {
				if format.URL != "" && format.VCodec != "none" {
					if format.Height > 0 && format.Height < minHeight {
						minHeight = format.Height
						selectedFormat = &format
					} else if selectedFormat == nil {
						selectedFormat = &format
					}
				}
			}
		case strings.Contains(quality, "height<="):
			// Extract target height from quality string like "best[height<=720]"
			var targetHeight int
			if _, err := fmt.Sscanf(quality, "best[height<=%d]", &targetHeight); err == nil {
				// Find best format that doesn't exceed target height
				bestHeight := 0
				for _, format := range info.Formats {
					if format.URL != "" && format.VCodec != "none" && format.Height <= targetHeight {
						if format.Height > bestHeight {
							bestHeight = format.Height
							selectedFormat = &format
						} else if selectedFormat == nil {
							selectedFormat = &format
						}
					}
				}
			}
		default:
			// For other qualities, try to find the best available
			maxHeight := 0
			for _, format := range info.Formats {
				if format.URL != "" && format.VCodec != "none" {
					if format.Height > maxHeight {
						maxHeight = format.Height
						selectedFormat = &format
					} else if selectedFormat == nil {
						selectedFormat = &format
					}
				}
			}
		}

		if selectedFormat != nil {
			videoURL = selectedFormat.URL
			fmt.Printf("DEBUG: Selected format with height %d for quality '%s': %s\n", selectedFormat.Height, quality, videoURL)
		}
	}

	// Fallback to info.URL if no format was selected
	if videoURL == "" {
		videoURL = info.URL
		fmt.Printf("DEBUG: Using fallback URL: %s\n", videoURL)
	}

	// Validate that we got a video URL
	if videoURL == "" {
		return nil, fmt.Errorf("no video URL found in yt-dlp output")
	}

	// Detect platform from URL
	platform := d.DetectPlatform(url)

	return &models.VideoResponse{
		VideoURL:    videoURL,
		Title:       info.Title,
		Platform:    platform,
		Quality:     quality,
		ProcessedAt: time.Now(),
		Metadata: map[string]string{
			"source":      url,
			"description": info.Description,
			"duration":    fmt.Sprintf("%.1f", info.Duration),
			"thumbnail":   info.Thumbnail,
		},
	}, nil
}

func (d *UniversalDownloader) GetAvailableQualities(url string) (*models.QualitiesResponse, error) {
	// Clean the URL by trimming whitespace
	url = strings.TrimSpace(url)

	// Check if it's a YouTube URL and return specific error
	if youtubePattern.MatchString(url) {
		return nil, fmt.Errorf("YouTube videos are not supported at this time")
	}

	if !d.ValidateURL(url) {
		return nil, fmt.Errorf("unsupported URL or platform")
	}

	// Check if yt-dlp is available
	if _, err := exec.LookPath(d.ytdlpPath); err != nil {
		return nil, fmt.Errorf("yt-dlp not found in PATH: %w", err)
	}

	// Get available formats using yt-dlp with --list-formats and --dump-json
	args := []string{
		"--dump-json",
		"--no-warnings",
		"--no-check-certificate",
		"--no-playlist",
		url,
	}

	fmt.Printf("DEBUG: Getting qualities with args: %v\n", args)

	cmd := exec.Command(d.ytdlpPath, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get available qualities: %w", err)
	}

	var info UniversalYtDlpInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, fmt.Errorf("failed to parse yt-dlp JSON: %w", err)
	}

	// Platform detection
	platform := d.DetectPlatform(url)

	// Parse available formats from yt-dlp output
	var qualities []models.QualityOption

	// Add standard quality options that work with yt-dlp
	qualities = append(qualities, models.QualityOption{
		Quality: "best",
		Label:   "Best Available",
	})

	// Parse specific formats if available
	if len(info.Formats) > 0 {
		seenQualities := make(map[string]bool)

		// Sort formats by height descending to get better quality first
		for _, format := range info.Formats {
			if format.Height > 0 && format.VCodec != "none" {
				qualityLabel := fmt.Sprintf("%dp", format.Height)
				qualityId := fmt.Sprintf("best[height<=%d]", format.Height)

				if !seenQualities[qualityLabel] {
					qualities = append(qualities, models.QualityOption{
						Quality: qualityId,
						Label:   qualityLabel,
					})
					seenQualities[qualityLabel] = true
				}
			}
		}

		// If no video formats found, try to add based on format IDs
		if len(seenQualities) == 0 {
			fmt.Printf("DEBUG: No video formats with height found, checking format IDs\n")
			for _, format := range info.Formats {
				fmt.Printf("DEBUG: Format ID: %s, Height: %d, VCodec: %s, ACodec: %s\n",
					format.FormatID, format.Height, format.VCodec, format.ACodec)

				// Add formats with specific IDs as quality options
				if format.FormatID != "" {
					label := format.FormatID
					if format.Height > 0 {
						label = fmt.Sprintf("%s (%dp)", format.FormatID, format.Height)
					}
					qualities = append(qualities, models.QualityOption{
						Quality: format.FormatID,
						Label:   label,
					})
				}
			}
		}
	} else {
		fmt.Printf("DEBUG: No formats found in yt-dlp output\n")
	}

	// Add worst quality option
	qualities = append(qualities, models.QualityOption{
		Quality: "worst",
		Label:   "Lowest Quality",
	})

	return &models.QualitiesResponse{
		Platform:           platform,
		AvailableQualities: qualities,
	}, nil
}
