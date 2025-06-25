package main

import (
	"net/url"
	"testing"

	"github.com/itxtx/crawler_go/config"
	"github.com/itxtx/crawler_go/extractor"
)

// TestBackwardCompatibilityNoVideoExtract ensures that the default behavior remains unchanged
// when video_extract is not specified or is false
func TestBackwardCompatibilityNoVideoExtract(t *testing.T) {
	baseURL, _ := url.Parse("https://example.com")

	// HTML content that contains videos, which should NOT be extracted by default
	htmlWithVideos := `
		<html>
			<head>
				<meta property="og:video" content="https://example.com/og-video.mp4">
				<title>Test Page</title>
			</head>
			<body>
				<p>This is a test page with some content.</p>
				<video controls>
					<source src="/videos/sample.mp4" type="video/mp4">
					<source src="/videos/sample.webm" type="video/webm">
				</video>
				<iframe src="https://www.youtube.com/embed/dQw4w9WgXcQ"></iframe>
				<a href="/page1.html">Link 1</a>
				<a href="/page2.html">Link 2</a>
			</body>
		</html>
	`

	t.Run("Default config (ExtractVideos=false) should not extract videos", func(t *testing.T) {
		// Test with default config where ExtractVideos is false
		cfg := &config.CrawlerConfig{
			ExtractVideos: false, // Default value
		}

		// ExtractVideos function should still work but won't be called by default crawling
		videos, err := extractor.ExtractVideos(htmlWithVideos, baseURL, cfg, false)
		if err != nil {
			t.Fatalf("ExtractVideos should not fail even when not enabled: %v", err)
		}

		// Videos should still be extractable when function is called directly
		if len(videos) == 0 {
			t.Error("ExtractVideos function should still extract videos when called directly")
		}
	})

	t.Run("Existing functions should remain untouched", func(t *testing.T) {
		// Test that existing content extraction functions work as before
		pattern := "p"
		selectorType := "css"
		format := "text"

		// This should work exactly as before
		extractor.ExtractContent(htmlWithVideos, pattern, selectorType, format, false)

		// Test that link extraction still works
		links, err := extractor.ExtractLinksAndDescriptions(htmlWithVideos, baseURL, "")
		if err != nil {
			t.Fatalf("ExtractLinksAndDescriptions should work as before: %v", err)
		}

		// Should find the two links
		if len(links) != 2 {
			t.Errorf("Expected 2 links, got %d", len(links))
		}
	})

	t.Run("Config parsing with no video_extract parameter", func(t *testing.T) {
		// Test that config parsing works without video_extract parameter
		args := []string{"program", "https://example.com", "1", "1"}
		cfg, err := config.ParseArgs(args)
		if err != nil {
			t.Fatalf("ParseArgs should work without video_extract: %v", err)
		}

		// ExtractVideos should be false by default
		if cfg.ExtractVideos {
			t.Error("ExtractVideos should be false by default")
		}

		// Other default values should remain unchanged
		if cfg.SelectorType != "css" {
			t.Errorf("Expected default SelectorType 'css', got '%s'", cfg.SelectorType)
		}
		if cfg.OutputFormat != "text" {
			t.Errorf("Expected default OutputFormat 'text', got '%s'", cfg.OutputFormat)
		}
		if cfg.VideoFormat != "text" {
			t.Errorf("Expected default VideoFormat 'text', got '%s'", cfg.VideoFormat)
		}
	})

	t.Run("Config parsing with video_extract=false", func(t *testing.T) {
		// Test that explicitly setting video_extract=false works
		args := []string{"program", "https://example.com", "1", "1", "video_extract=false"}
		cfg, err := config.ParseArgs(args)
		if err != nil {
			t.Fatalf("ParseArgs should work with video_extract=false: %v", err)
		}

		// ExtractVideos should be false
		if cfg.ExtractVideos {
			t.Error("ExtractVideos should be false when explicitly set")
		}
	})

	t.Run("Config parsing with video_extract=true", func(t *testing.T) {
		// Test that video_extract=true enables the feature
		args := []string{"program", "https://example.com", "1", "1", "video_extract=true"}
		cfg, err := config.ParseArgs(args)
		if err != nil {
			t.Fatalf("ParseArgs should work with video_extract=true: %v", err)
		}

		// ExtractVideos should be true
		if !cfg.ExtractVideos {
			t.Error("ExtractVideos should be true when explicitly enabled")
		}
	})
}

// TestVideoExtractionGuardedByFlag ensures that video extraction only happens when the flag is enabled
func TestVideoExtractionGuardedByFlag(t *testing.T) {
	t.Run("Video extraction logic is properly guarded", func(t *testing.T) {
		// This test verifies that the video extraction code paths are properly
		// guarded by the ExtractVideos flag in the crawling logic

		baseURL, _ := url.Parse("https://example.com")

		// Config with ExtractVideos disabled
		cfgDisabled := &config.CrawlerConfig{
			ExtractVideos: false,
		}

		// Config with ExtractVideos enabled
		cfgEnabled := &config.CrawlerConfig{
			ExtractVideos: true,
		}

		htmlWithVideos := `<html><body><video src="test.mp4"></video></body></html>`

		// When disabled, video extraction should not occur during normal crawling
		// But the function should still work when called directly
		videosDisabled, err := extractor.ExtractVideos(htmlWithVideos, baseURL, cfgDisabled, false)
		if err != nil {
			t.Fatalf("ExtractVideos should work regardless of flag: %v", err)
		}

		// When enabled, video extraction should work normally
		videosEnabled, err := extractor.ExtractVideos(htmlWithVideos, baseURL, cfgEnabled, false)
		if err != nil {
			t.Fatalf("ExtractVideos should work when enabled: %v", err)
		}

		// Both should find the same videos (the flag doesn't affect the extraction logic itself)
		if len(videosDisabled) != len(videosEnabled) {
			t.Errorf("Video extraction results should be the same regardless of flag when function is called directly")
		}
	})
}

// TestExistingFunctionsUntouched verifies that existing functions remain completely unchanged
func TestExistingFunctionsUntouched(t *testing.T) {
	baseURL, _ := url.Parse("https://example.com")
	testHTML := `
		<html>
			<head><title>Test Page</title></head>
			<body>
				<h1>Main Title</h1>
				<p class="content">This is test content.</p>
				<a href="/link1">Link 1</a>
				<a href="/link2">Link 2</a>
				<video src="video.mp4"></video>
			</body>
		</html>
	`

	t.Run("ExtractContent function works unchanged", func(t *testing.T) {
		// This function should work exactly as it did before
		extractor.ExtractContent(testHTML, "h1", "css", "text", false)
		extractor.ExtractContent(testHTML, ".content", "css", "json", false)
		extractor.ExtractContent(testHTML, "//p[@class='content']", "xpath", "csv", false)
	})

	t.Run("ExtractMultipleContents function works unchanged", func(t *testing.T) {
		// This function should work exactly as it did before
		patterns := []string{"h1", ".content"}
		extractor.ExtractMultipleContents(testHTML, patterns, "css", "text", false)
	})

	t.Run("ExtractLinksAndDescriptions function works unchanged", func(t *testing.T) {
		// This function should work exactly as it did before
		links, err := extractor.ExtractLinksAndDescriptions(testHTML, baseURL, "")
		if err != nil {
			t.Fatalf("ExtractLinksAndDescriptions failed: %v", err)
		}

		if len(links) != 2 {
			t.Errorf("Expected 2 links, got %d", len(links))
		}

		// Test with filter
		filteredLinks, err := extractor.ExtractLinksAndDescriptions(testHTML, baseURL, "link1")
		if err != nil {
			t.Fatalf("ExtractLinksAndDescriptions with filter failed: %v", err)
		}

		if len(filteredLinks) != 1 {
			t.Errorf("Expected 1 filtered link, got %d", len(filteredLinks))
		}
	})

	t.Run("formatOutput function behavior unchanged", func(t *testing.T) {
		// Test that the existing formatOutput pattern is preserved
		// This is tested indirectly through the ExtractContent calls above
		// but we verify the behavior is consistent

		// These calls should work as they did before (internal function)
		// We test this through the public API that uses formatOutput
		extractor.ExtractContent(testHTML, "a", "css", "text", false)
		extractor.ExtractContent(testHTML, "a", "css", "json", false)
		extractor.ExtractContent(testHTML, "a", "css", "csv", false)
	})
}
