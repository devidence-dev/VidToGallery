package models

import "time"

type VideoRequest struct {
	URL     string `json:"url" validate:"required"`
	Quality string `json:"quality,omitempty"`
}

type VideoResponse struct {
	VideoURL    string            `json:"video_url"`
	Title       string            `json:"title,omitempty"`
	Duration    int               `json:"duration,omitempty"`
	Platform    string            `json:"platform"`
	Quality     string            `json:"quality"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	ProcessedAt time.Time         `json:"processed_at"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}
