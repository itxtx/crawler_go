package main

import (
	"fmt"
	"net/url"

	"github.com/itxtx/crawler_go/extractor"
)

func main() {
	// Example HTML content with various types of video embeds
	htmlContent := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Video Examples Page</title>
		<meta property="og:video" content="https://example.com/og-video.mp4">
		<meta property="og:title" content="Open Graph Video">
		<meta property="og:image" content="https://example.com/og-thumb.jpg">
		
		<script type="application/ld+json">
		{
			"@context": "https://schema.org",
			"@type": "VideoObject",
			"name": "Schema.org Video",
			"contentUrl": "https://example.com/schema-video.mp4",
			"thumbnailUrl": "https://example.com/schema-thumb.jpg",
			"width": 1920,
			"height": 1080
		}
		</script>
	</head>
	<body>
		<h1>Video Examples</h1>
		
		<!-- HTML5 Video -->
		<video controls width="640" height="360" poster="https://example.com/poster.jpg">
			<source src="/videos/sample.mp4" type="video/mp4">
			<source src="/videos/sample.webm" type="video/webm">
			Your browser does not support the video tag.
		</video>
		
		<!-- YouTube Embed -->
		<iframe width="560" height="315" 
			src="https://www.youtube.com/embed/dQw4w9WgXcQ" 
			frameborder="0" 
			allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" 
			allowfullscreen>
		</iframe>
		
		<!-- Vimeo Embed -->
		<iframe src="https://player.vimeo.com/video/123456789" 
			width="640" height="360" frameborder="0" 
			allow="autoplay; fullscreen; picture-in-picture" 
			allowfullscreen>
		</iframe>
		
		<!-- Direct video link in text -->
		<p>Check out this amazing video: https://example.com/direct-video.mp4</p>
		
		<!-- Custom data attributes -->
		<div class="video-container" 
			data-video-src="https://example.com/custom-video.mp4"
			data-poster="https://example.com/custom-thumb.jpg"
			data-title="Custom Video Player">
			Click to play video
		</div>
		
		<!-- Another custom data attribute variant -->
		<button data-video-url="/relative/video.mov" title="Play Button Video">
			Play Video
		</button>
	</body>
	</html>
	`

	// Base URL for resolving relative URLs
	baseURL, err := url.Parse("https://example.com")
	if err != nil {
		fmt.Printf("Error parsing base URL: %v\n", err)
		return
	}

	// Extract videos using all detection strategies
	// Pass nil for config since we're not using any special configuration
	// Pass false for printContent since we want to handle the output ourselves
	videos, err := extractor.ExtractVideos(htmlContent, baseURL, nil, false)
	if err != nil {
		fmt.Printf("Error extracting videos: %v\n", err)
		return
	}

	// Display results
	fmt.Printf("Found %d videos:\n\n", len(videos))

	for i, video := range videos {
		fmt.Printf("Video %d:\n", i+1)
		fmt.Printf("  URL: %s\n", video.URL)
		fmt.Printf("  Title: %s\n", video.Title)
		fmt.Printf("  Platform: %s\n", video.Platform)
		fmt.Printf("  Thumbnail: %s\n", video.Thumbnail)
		if video.Width > 0 && video.Height > 0 {
			fmt.Printf("  Dimensions: %dx%d\n", video.Width, video.Height)
		}
		fmt.Println()
	}
}
