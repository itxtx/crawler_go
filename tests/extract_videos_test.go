package main

import (
	"net/url"
	"testing"

	"github.com/itxtx/crawler_go/config"
	"github.com/itxtx/crawler_go/extractor"
)

func TestExtractVideos(t *testing.T) {
	baseURL, _ := url.Parse("https://example.com")

	testCases := []struct {
		name     string
		html     string
		expected int
	}{
		{
			name: "HTML5 video with source tag",
			html: `
				<html>
					<body>
						<video controls>
							<source src="/videos/sample.mp4" type="video/mp4">
							<source src="/videos/sample.webm" type="video/webm">
						</video>
					</body>
				</html>
			`,
			expected: 2,
		},
		{
			name: "YouTube iframe embed",
			html: `
				<html>
					<body>
						<iframe width="560" height="315" src="https://www.youtube.com/embed/dQw4w9WgXcQ" frameborder="0"></iframe>
					</body>
				</html>
			`,
			expected: 1,
		},
		{
			name: "Direct video URL in text",
			html: `
				<html>
					<body>
						<p>Check out this video: https://example.com/video.mp4</p>
					</body>
				</html>
			`,
			expected: 1,
		},
		{
			name: "Open Graph video meta tag",
			html: `
				<html>
					<head>
						<meta property="og:video" content="https://example.com/video.mp4">
						<meta property="og:title" content="Test Video">
					</head>
					<body></body>
				</html>
			`,
			expected: 1,
		},
		{
			name: "Custom data attribute",
			html: `
				<html>
					<body>
						<div data-video-src="https://example.com/video.mp4" data-title="Custom Video"></div>
					</body>
				</html>
			`,
			expected: 1,
		},
		{
			name: "JSON-LD VideoObject",
			html: `
				<html>
					<head>
						<script type="application/ld+json">
						{
							"@context": "https://schema.org",
							"@type": "VideoObject",
							"name": "Test Video",
							"contentUrl": "https://example.com/video.mp4",
							"thumbnailUrl": "https://example.com/thumb.jpg"
						}
						</script>
					</head>
					<body></body>
				</html>
			`,
			expected: 1,
		},
		{
			name: "Multiple video sources with deduplication",
			html: `
				<html>
					<head>
						<meta property="og:video" content="https://example.com/video.mp4">
					</head>
					<body>
						<video src="https://example.com/video.mp4"></video>
						<p>Watch: https://example.com/video.mp4</p>
						<iframe src="https://www.youtube.com/embed/abc123"></iframe>
					</body>
				</html>
			`,
			expected: 2, // One deduplicated MP4 and one YouTube video
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			videos, err := extractor.ExtractVideos(tc.html, baseURL, nil, false)
			if err != nil {
				t.Fatalf("ExtractVideos failed: %v", err)
			}

			if len(videos) != tc.expected {
				t.Errorf("Expected %d videos, got %d", tc.expected, len(videos))
				for i, video := range videos {
					t.Logf("Video %d: URL=%s, Platform=%s, Title=%s", i+1, video.URL, video.Platform, video.Title)
				}
			}

			// Check that all videos have valid URLs
			for _, video := range videos {
				if video.URL == "" {
					t.Error("Found video with empty URL")
				}
				if video.Platform == "" {
					t.Error("Found video with empty platform")
				}
			}
		})
	}
}

func TestExtractVideosEdgeCases(t *testing.T) {
	baseURL, _ := url.Parse("https://example.com")

	t.Run("Empty HTML", func(t *testing.T) {
		videos, err := extractor.ExtractVideos("", baseURL, nil, false)
		if err != nil {
			t.Fatalf("ExtractVideos failed: %v", err)
		}
		if len(videos) != 0 {
			t.Errorf("Expected 0 videos for empty HTML, got %d", len(videos))
		}
	})

	t.Run("Invalid HTML", func(t *testing.T) {
		videos, err := extractor.ExtractVideos("<html><body><video><source", baseURL, nil, false)
		if err != nil {
			t.Fatalf("ExtractVideos failed: %v", err)
		}
		// Should handle gracefully even with malformed HTML
		// (Note: This is just checking that we get a valid slice back)
		if videos == nil {
			t.Error("Should handle malformed HTML gracefully and return non-nil slice")
		}
	})

	t.Run("Relative URLs", func(t *testing.T) {
		html := `<html><body><video src="/relative/video.mp4"></video></body></html>`
		videos, err := extractor.ExtractVideos(html, baseURL, nil, false)
		if err != nil {
			t.Fatalf("ExtractVideos failed: %v", err)
		}
		if len(videos) != 1 {
			t.Fatalf("Expected 1 video, got %d", len(videos))
		}
		expected := "https://example.com/relative/video.mp4"
		if videos[0].URL != expected {
			t.Errorf("Expected URL %s, got %s", expected, videos[0].URL)
		}
	})
}

func TestExtractVideosWithCustomSelectors(t *testing.T) {
	baseURL, _ := url.Parse("https://example.com")

	t.Run("CSS Selector for custom video elements", func(t *testing.T) {
		html := `
			<html>
				<body>
					<div class="video-container">
						<a href="https://example.com/custom-video.mp4" class="video-link">Watch Video</a>
					</div>
					<span data-video-url="https://example.com/another-video.mp4">Another Video</span>
				</body>
			</html>
		`
		cfg := &config.CrawlerConfig{
			Selectors:    []string{".video-link", "[data-video-url]"},
			SelectorType: "css",
		}

		videos, err := extractor.ExtractVideos(html, baseURL, cfg, false)
		if err != nil {
			t.Fatalf("ExtractVideos failed: %v", err)
		}

		// Should find 2 videos: one from href attribute and one from data-video-url
		if len(videos) != 2 {
			t.Errorf("Expected 2 videos with custom CSS selectors, got %d", len(videos))
			for i, video := range videos {
				t.Logf("Video %d: URL=%s, Platform=%s", i+1, video.URL, video.Platform)
			}
		}

		// Verify video URLs
		expectedURLs := []string{
			"https://example.com/custom-video.mp4",
			"https://example.com/another-video.mp4",
		}
		foundURLs := make(map[string]bool)
		for _, video := range videos {
			foundURLs[video.URL] = true
			if video.Platform != "custom" {
				t.Errorf("Expected platform 'custom', got '%s'", video.Platform)
			}
		}
		for _, expectedURL := range expectedURLs {
			if !foundURLs[expectedURL] {
				t.Errorf("Expected to find video URL %s", expectedURL)
			}
		}
	})

	t.Run("XPath Selector for custom video elements", func(t *testing.T) {
		html := `
			<html>
				<body>
					<div class="media">
						<video src="/xpath-video1.mp4"></video>
						<source src="/xpath-video2.webm" />
					</div>
				</body>
			</html>
		`
		cfg := &config.CrawlerConfig{
			Selectors:    []string{"//div[@class='media']//video", "//div[@class='media']//source"},
			SelectorType: "xpath",
		}

		videos, err := extractor.ExtractVideos(html, baseURL, cfg, false)
		if err != nil {
			t.Fatalf("ExtractVideos failed: %v", err)
		}

		// Should find videos from XPath selectors
		if len(videos) < 2 {
			t.Errorf("Expected at least 2 videos with XPath selectors, got %d", len(videos))
			for i, video := range videos {
				t.Logf("Video %d: URL=%s, Platform=%s", i+1, video.URL, video.Platform)
			}
		}
	})

	t.Run("Regex Selector for custom video patterns", func(t *testing.T) {
		html := `
			<html>
				<body>
					<p>Check out this amazing video: [VIDEO:https://example.com/regex-video1.mp4]</p>
					<div>Another video link: [VIDEO:https://example.com/regex-video2.webm]</div>
					<span>Not a video: [AUDIO:https://example.com/audio.mp3]</span>
				</body>
			</html>
		`
		cfg := &config.CrawlerConfig{
			Selectors:    []string{`\[VIDEO:(https?://[^\]]+)\]`},
			SelectorType: "regex",
		}

		videos, err := extractor.ExtractVideos(html, baseURL, cfg, false)
		if err != nil {
			t.Fatalf("ExtractVideos failed: %v", err)
		}

		// Should find 2 videos from regex pattern
		if len(videos) != 2 {
			t.Errorf("Expected 2 videos with regex selector, got %d", len(videos))
			for i, video := range videos {
				t.Logf("Video %d: URL=%s, Platform=%s", i+1, video.URL, video.Platform)
			}
		}

		// Verify that custom platform is set
		for _, video := range videos {
			if video.Platform != "custom" {
				t.Errorf("Expected platform 'custom', got '%s'", video.Platform)
			}
		}
	})

	t.Run("Custom selectors work with standard selector types", func(t *testing.T) {
		html := `
			<html>
				<body>
					<a href="https://example.com/video.mp4" class="video-link">Video</a>
				</body>
			</html>
		`
		// Config with selectors using standard CSS selector type
		cfg := &config.CrawlerConfig{
			Selectors:    []string{".video-link"},
			SelectorType: "css", // Standard CSS selector type
		}

		videos, err := extractor.ExtractVideos(html, baseURL, cfg, false)
		if err != nil {
			t.Fatalf("ExtractVideos failed: %v", err)
		}

		// Should find 1 video with custom platform
		if len(videos) != 1 {
			t.Errorf("Expected 1 video with CSS selector, got %d", len(videos))
		}

		// Check that video has custom platform
		for _, video := range videos {
			if video.Platform != "custom" {
				t.Errorf("Expected platform 'custom', got '%s'", video.Platform)
			}
		}
	})

	t.Run("Mixed custom selectors with existing video detection", func(t *testing.T) {
		html := `
			<html>
				<head>
					<meta property="og:video" content="https://example.com/og-video.mp4">
				</head>
				<body>
					<video src="/html5-video.mp4"></video>
					<iframe src="https://www.youtube.com/embed/xyz123"></iframe>
					<a href="https://example.com/custom-video.mp4" class="custom-video">Custom Video</a>
				</body>
			</html>
		`
		cfg := &config.CrawlerConfig{
			Selectors:    []string{".custom-video"},
			SelectorType: "css", // Standard CSS selector type
		}

		videos, err := extractor.ExtractVideos(html, baseURL, cfg, false)
		if err != nil {
			t.Fatalf("ExtractVideos failed: %v", err)
		}

		// Should find videos from both standard detection and custom selectors
		if len(videos) < 3 {
			t.Errorf("Expected at least 3 videos (standard + custom), got %d", len(videos))
			for i, video := range videos {
				t.Logf("Video %d: URL=%s, Platform=%s", i+1, video.URL, video.Platform)
			}
		}

		// Verify that we have videos from different platforms
		platforms := make(map[string]bool)
		for _, video := range videos {
			platforms[video.Platform] = true
		}

		// Check for key platforms (some may overlap or be detected differently)
		if !platforms["html5"] && !platforms["direct"] {
			t.Error("Expected to find HTML5 or direct video platform")
		}
		if !platforms["youtube"] {
			t.Error("Expected to find YouTube platform")
		}
		if !platforms["custom"] {
			t.Error("Expected to find custom platform")
		}
	})

	t.Run("No custom selectors when config is nil", func(t *testing.T) {
		html := `
			<html>
				<body>
					<a href="https://example.com/video.mp4" class="video-link">Video</a>
					<video src="/standard-video.mp4"></video>
				</body>
			</html>
		`

		videos, err := extractor.ExtractVideos(html, baseURL, nil, false)
		if err != nil {
			t.Fatalf("ExtractVideos failed: %v", err)
		}

		// Should find videos through standard detection (HTML5 video tag + direct URL detection)
		if len(videos) < 1 {
			t.Errorf("Expected at least 1 video through standard detection, got %d", len(videos))
			for i, video := range videos {
				t.Logf("Video %d: URL=%s, Platform=%s", i+1, video.URL, video.Platform)
			}
		}

		for _, video := range videos {
			if video.Platform == "custom" {
				t.Error("Should not find custom platform videos when config is nil")
			}
		}
	})
}
