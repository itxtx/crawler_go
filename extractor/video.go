package extractor

import (
	"encoding/json"
	"fmt"
	"strings"
)

// VideoInfo represents information about a video from various platforms
type VideoInfo struct {
	URL       string // The video URL
	Title     string // The video title
	Thumbnail string // The thumbnail image URL
	Platform  string // youtube, vimeo, html5, direct, custom
	Width     int    // Video width in pixels
	Height    int    // Video height in pixels
}

// FormatVideoOutput formats a slice of VideoInfo structs according to the specified format
// Supports "json", "csv", and "text" (default) formats, mirroring the existing formatOutput function
func FormatVideoOutput(videos []VideoInfo, format string) string {
	switch format {
	case "json":
		jsonData, err := json.MarshalIndent(videos, "", "  ")
		if err != nil {
			return fmt.Sprintf("Error marshaling JSON: %v", err)
		}
		return string(jsonData)
	case "csv":
		if len(videos) == 0 {
			return ""
		}
		// CSV header
		lines := []string{"URL,Title,Platform,Thumbnail"}
		// CSV data rows
		for _, video := range videos {
			// Escape commas and quotes in CSV fields
			url := escapeCSVField(video.URL)
			title := escapeCSVField(video.Title)
			platform := escapeCSVField(video.Platform)
			thumbnail := escapeCSVField(video.Thumbnail)
			line := fmt.Sprintf("%s,%s,%s,%s", url, title, platform, thumbnail)
			lines = append(lines, line)
		}
		return strings.Join(lines, "\n")
	default:
		// Text format (default)
		if len(videos) == 0 {
			return "No videos found"
		}
		var lines []string
		for i, video := range videos {
			lines = append(lines, fmt.Sprintf("Video %d:", i+1))
			lines = append(lines, fmt.Sprintf("  URL: %s", video.URL))
			lines = append(lines, fmt.Sprintf("  Title: %s", video.Title))
			lines = append(lines, fmt.Sprintf("  Thumbnail: %s", video.Thumbnail))
			lines = append(lines, fmt.Sprintf("  Platform: %s", video.Platform))
			lines = append(lines, fmt.Sprintf("  Dimensions: %dx%d", video.Width, video.Height))
			lines = append(lines, "") // Empty line between videos
		}
		return strings.Join(lines, "\n")
	}
}

// escapeCSVField escapes a field for CSV output by wrapping in quotes if it contains commas, quotes, or newlines
func escapeCSVField(field string) string {
	if strings.Contains(field, ",") || strings.Contains(field, "\"") || strings.Contains(field, "\n") {
		// Escape quotes by doubling them and wrap the whole field in quotes
		escaped := strings.ReplaceAll(field, "\"", "\"\"")
		return fmt.Sprintf("\"%s\"", escaped)
	}
	return field
}
