package quality

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"vidtogallery/internal/models"
)

// Manager handles quality selection and processing logic
type Manager struct{}

// NewManager creates a new quality manager
func NewManager() *Manager {
	return &Manager{}
}

// SelectBestQuality returns the highest quality option from a list
func (m *Manager) SelectBestQuality(qualities []models.QualityOption) *models.QualityOption {
	if len(qualities) == 0 {
		return nil
	}

	bestQuality := qualities[0]
	bestResolution := m.calculateResolution(bestQuality.Width, bestQuality.Height)

	for _, quality := range qualities[1:] {
		resolution := m.calculateResolution(quality.Width, quality.Height)
		if resolution > bestResolution {
			bestQuality = quality
			bestResolution = resolution
		}
	}

	return &bestQuality
}

// SelectQualityByPreference selects quality based on user preference
func (m *Manager) SelectQualityByPreference(qualities []models.QualityOption, preference string) *models.QualityOption {
	if len(qualities) == 0 {
		return nil
	}

	// If preference is "best" or empty, return best quality
	if preference == "best" || preference == "" {
		return m.SelectBestQuality(qualities)
	}

	// Try to find exact match
	for _, quality := range qualities {
		if quality.Quality == preference {
			return &quality
		}
	}

	// Fallback to best quality
	return m.SelectBestQuality(qualities)
}

// SortQualitiesByResolution sorts qualities from highest to lowest resolution
func (m *Manager) SortQualitiesByResolution(qualities []models.QualityOption) []models.QualityOption {
	sorted := make([]models.QualityOption, len(qualities))
	copy(sorted, qualities)

	sort.Slice(sorted, func(i, j int) bool {
		resI := m.calculateResolution(sorted[i].Width, sorted[i].Height)
		resJ := m.calculateResolution(sorted[j].Width, sorted[j].Height)
		return resI > resJ
	})

	return sorted
}

// ParseDimensions extracts width and height from quality string like "1920x1080"
func (m *Manager) ParseDimensions(quality string) (width, height int, err error) {
	parts := strings.Split(quality, "x")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid quality format: %s", quality)
	}

	width, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid width in quality: %s", quality)
	}

	height, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid height in quality: %s", quality)
	}

	return width, height, nil
}

// FormatQualityLabel creates a user-friendly label from dimensions
func (m *Manager) FormatQualityLabel(width, height int) string {
	switch {
	case height >= 2160:
		return "4K"
	case height >= 1440:
		return "1440p"
	case height >= 1080:
		return "1080p"
	case height >= 720:
		return "720p"
	case height >= 480:
		return "480p"
	case height >= 360:
		return "360p"
	default:
		return fmt.Sprintf("%dp", height)
	}
}

// calculateResolution returns total pixel count for comparison
func (m *Manager) calculateResolution(width, height int) int {
	return width * height
}
