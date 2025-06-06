package models

import "time"

type VideoRequest struct {
	URL     string `json:"url" validate:"required"`
	Quality string `json:"quality,omitempty"`
}

type VideoResponse struct {
	VideoURL           string            `json:"video_url"`
	Title              string            `json:"title,omitempty"`
	Duration           int               `json:"duration,omitempty"`
	Platform           string            `json:"platform"`
	Quality            string            `json:"quality"`
	Metadata           map[string]string `json:"metadata,omitempty"`
	ProcessedAt        time.Time         `json:"processed_at"`
	AvailableQualities []QualityOption   `json:"available_qualities,omitempty"`
}

type QualityOption struct {
	Quality  string `json:"quality"`
	Label    string `json:"label"`
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
	VideoURL string `json:"video_url"`
}

type QualitiesResponse struct {
	AvailableQualities []QualityOption `json:"available_qualities"`
	Platform           string          `json:"platform"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

type QualityRequest struct {
	URL string `json:"url" validate:"required"`
}

type ProxyDownloadRequest struct {
	VideoURL string `json:"video_url" validate:"required"`
}

type ProxyDownloadResponse struct {
	Data []byte `json:"-"`
}
