package main

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/itxtx/crawler_go/config"
	"github.com/itxtx/crawler_go/extractor"
)

func TestExtractionErrorTypes(t *testing.T) {
	t.Run("NewMalformedURLError", func(t *testing.T) {
		err := extractor.NewMalformedURLError("invalid://url[", errors.New("parse error"))
		if err.Type != extractor.ErrorTypeMalformedURL {
			t.Errorf("Expected error type %s, got %s", extractor.ErrorTypeMalformedURL, err.Type)
		}
		if !strings.Contains(err.Error(), "invalid://url[") {
			t.Error("Error message should contain the source URL")
		}
	})

	t.Run("NewMissingSrcError", func(t *testing.T) {
		err := extractor.NewMissingSrcError("video", "test context")
		if err.Type != extractor.ErrorTypeMissingSrc {
			t.Errorf("Expected error type %s, got %s", extractor.ErrorTypeMissingSrc, err.Type)
		}
		if !strings.Contains(err.Error(), "video element") {
			t.Error("Error message should mention the element type")
		}
	})

	t.Run("NewJSONLDParseError", func(t *testing.T) {
		invalidJSON := `{"@type": "VideoObject", "name": "Test", invalid}`
		err := extractor.NewJSONLDParseError(invalidJSON, errors.New("json parse error"))
		if err.Type != extractor.ErrorTypeJSONLDParse {
			t.Errorf("Expected error type %s, got %s", extractor.ErrorTypeJSONLDParse, err.Type)
		}
		if !strings.Contains(err.Error(), "Failed to parse JSON-LD") {
			t.Error("Error message should mention JSON-LD parsing")
		}
	})

	t.Run("Error_Unwrap", func(t *testing.T) {
		originalErr := errors.New("original error")
		extractionErr := extractor.NewMalformedURLError("test", originalErr)
		
		if extractionErr.Unwrap() != originalErr {
			t.Error("Unwrap should return the original error")
		}
	})

	t.Run("Error_Is", func(t *testing.T) {
		err1 := extractor.NewMalformedURLError("test1", nil)
		err2 := extractor.NewMalformedURLError("test2", nil)
		err3 := extractor.NewMissingSrcError("video", "context")

		if !err1.Is(err2) {
			t.Error("Two errors of the same type should be equal with Is()")
		}
		if err1.Is(err3) {
			t.Error("Two errors of different types should not be equal with Is()")
		}
	})
}

func TestValidateURL(t *testing.T) {
	baseURL, _ := url.Parse("https://example.com")

	testCases := []struct {
		name        string
		rawURL      string
		base        *url.URL
		expectError bool
		errorType   extractor.ErrorType
	}{
		{
			name:        "Empty URL",
			rawURL:      "",
			base:        baseURL,
			expectError: true,
			errorType:   extractor.ErrorTypeMissingSrc,
		},
		{
			name:        "Invalid URL",
			rawURL:      "invalid://url[bad",
			base:        baseURL,
			expectError: true,
			errorType:   extractor.ErrorTypeMalformedURL,
		},
		{
			name:        "Valid absolute URL",
			rawURL:      "https://test.com/video.mp4",
			base:        baseURL,
			expectError: false,
		},
		{
			name:        "Valid relative URL with base",
			rawURL:      "/video.mp4",
			base:        baseURL,
			expectError: false,
		},
		{
			name:        "Relative URL without base",
			rawURL:      "/video.mp4",
			base:        nil,
			expectError: true,
			errorType:   extractor.ErrorTypeMalformedURL,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := extractor.ValidateURL(tc.rawURL, tc.base)

			if tc.expectError {
				if err == nil {
					t.Error("Expected an error but got none")
					return
				}
				if extractionErr, ok := err.(*extractor.ExtractionError); ok {
					if extractionErr.Type != tc.errorType {
						t.Errorf("Expected error type %s, got %s", tc.errorType, extractionErr.Type)
					}
				} else {
					t.Error("Expected ExtractionError type")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if result == "" {
					t.Error("Expected a valid URL result")
				}
			}
		})
	}
}

func TestSafeURLNormalize(t *testing.T) {
	baseURL, _ := url.Parse("https://example.com")

	t.Run("Valid URL normalization", func(t *testing.T) {
		result, err := extractor.SafeURLNormalize("/test.mp4", baseURL)
		if err != nil {
			t.Errorf("Expected no error but got: %v", err)
		}
		expected := "https://example.com/test.mp4"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("Invalid URL normalization", func(t *testing.T) {
		result, err := extractor.SafeURLNormalize("", baseURL)
		if err == nil {
			t.Error("Expected an error for empty URL")
		}
		if result != "" {
			t.Error("Expected empty result for invalid URL")
		}
		if err.Type != extractor.ErrorTypeMissingSrc {
			t.Errorf("Expected %s error type, got %s", extractor.ErrorTypeMissingSrc, err.Type)
		}
	})
}

func TestExtractVideosErrorHandling(t *testing.T) {
	baseURL, _ := url.Parse("https://example.com")

	t.Run("Empty HTML content", func(t *testing.T) {
		videos, err := extractor.ExtractVideos("", baseURL, nil, false)
		if err != nil {
			t.Fatalf("ExtractVideos should handle empty HTML gracefully: %v", err)
		}
		if len(videos) != 0 {
			t.Error("Expected no videos from empty HTML")
		}
	})

	t.Run("Malformed HTML content", func(t *testing.T) {
		malformedHTML := `<html><body><video><source src="/video.mp4"`
		videos, err := extractor.ExtractVideos(malformedHTML, baseURL, nil, false)
		// Should not crash, might find some content
		if err != nil {
			t.Fatalf("ExtractVideos should handle malformed HTML gracefully: %v", err)
		}
		// The exact number may vary based on how the parser handles malformed HTML
		t.Logf("Found %d videos in malformed HTML", len(videos))
	})

	t.Run("Invalid JSON-LD content", func(t *testing.T) {
		htmlWithInvalidJSON := `
			<html>
				<head>
					<script type="application/ld+json">
					{
						"@context": "https://schema.org",
						"@type": "VideoObject",
						"name": "Test Video",
						"contentUrl": "https://example.com/video.mp4",
						invalid json here
					}
					</script>
				</head>
				<body></body>
			</html>
		`
		
		// Set log level to capture errors during test
		extractor.SetLogLevel(extractor.LogLevelDebug)
		
		videos, err := extractor.ExtractVideos(htmlWithInvalidJSON, baseURL, nil, false)
		if err != nil {
			t.Fatalf("ExtractVideos should handle invalid JSON-LD gracefully: %v", err)
		}
		
		// Should not find the video from the invalid JSON-LD, but should not crash
		t.Logf("Found %d videos despite invalid JSON-LD", len(videos))
	})

	t.Run("Videos with missing src attributes", func(t *testing.T) {
		htmlWithMissingSrc := `
			<html>
				<body>
					<video controls>
						<!-- source tag without src -->
						<source type="video/mp4">
					</video>
					<iframe frameborder="0">
						<!-- iframe without src -->
					</iframe>
				</body>
			</html>
		`
		
		videos, err := extractor.ExtractVideos(htmlWithMissingSrc, baseURL, nil, false)
		if err != nil {
			t.Fatalf("ExtractVideos should handle missing src gracefully: %v", err)
		}
		
		// Should not find any videos due to missing src attributes
		if len(videos) != 0 {
			t.Errorf("Expected 0 videos with missing src, got %d", len(videos))
		}
	})

	t.Run("Invalid custom XPath selector", func(t *testing.T) {
		html := `<html><body><video src="/test.mp4"></video></body></html>`
		cfg := &config.CrawlerConfig{
			Selectors:    []string{"invalid-xpath[[["},
			SelectorType: "xpath",
		}
		
		// Set log level to capture errors during test
		extractor.SetLogLevel(extractor.LogLevelDebug)
		
		videos, err := extractor.ExtractVideos(html, baseURL, cfg, false)
		if err != nil {
			t.Fatalf("ExtractVideos should handle invalid XPath gracefully: %v", err)
		}
		
		// Should still find the video from standard detection
		if len(videos) == 0 {
			t.Error("Should still find videos through standard detection")
		}
	})

	t.Run("Invalid custom regex selector", func(t *testing.T) {
		html := `<html><body>Check out [VIDEO:https://example.com/test.mp4]</body></html>`
		cfg := &config.CrawlerConfig{
			Selectors:    []string{"[invalid-regex(("},
			SelectorType: "regex",
		}
		
		// Set log level to capture errors during test
		extractor.SetLogLevel(extractor.LogLevelDebug)
		
		videos, err := extractor.ExtractVideos(html, baseURL, cfg, false)
		if err != nil {
			t.Fatalf("ExtractVideos should handle invalid regex gracefully: %v", err)
		}
		
		// Should still find videos through standard detection if any
		t.Logf("Found %d videos despite invalid regex", len(videos))
	})

	t.Run("HTML parsing errors for XPath", func(t *testing.T) {
		// This is a bit tricky to test as goquery is quite forgiving
		// We'll test with a minimal configuration that might cause issues
		html := "not valid html at all"
		cfg := &config.CrawlerConfig{
			Selectors:    []string{"//video"},
			SelectorType: "xpath",
		}
		
		videos, err := extractor.ExtractVideos(html, baseURL, cfg, false)
		if err != nil {
			t.Fatalf("ExtractVideos should handle HTML parsing errors gracefully: %v", err)
		}
		
		t.Logf("Found %d videos from invalid HTML", len(videos))
	})
}

func TestJSONLDErrorHandling(t *testing.T) {
	baseURL, _ := url.Parse("https://example.com")

	testCases := []struct {
		name         string
		jsonContent  string
		expectVideos int
	}{
		{
			name: "Empty JSON-LD script",
			jsonContent: `
				<script type="application/ld+json"></script>
			`,
			expectVideos: 0,
		},
		{
			name: "Whitespace-only JSON-LD script", 
			jsonContent: `
				<script type="application/ld+json">   
				
				</script>
			`,
			expectVideos: 0,
		},
		{
			name: "Invalid JSON syntax",
			jsonContent: `
				<script type="application/ld+json">
				{ "name": "test", invalid }
				</script>
			`,
			expectVideos: 0,
		},
		{
			name: "Valid JSON but not VideoObject",
			jsonContent: `
				<script type="application/ld+json">
				{
					"@context": "https://schema.org",
					"@type": "Article",
					"name": "Test Article"
				}
				</script>
			`,
			expectVideos: 0,
		},
		{
			name: "Valid VideoObject with malformed URL",
			jsonContent: `
				<script type="application/ld+json">
				{
					"@context": "https://schema.org", 
					"@type": "VideoObject",
					"name": "Test Video",
					"contentUrl": "invalid://url["
				}
				</script>
			`,
			expectVideos: 0, // Should be filtered out due to invalid URL
		},
		{
			name: "VideoObject with missing contentUrl",
			jsonContent: `
				<script type="application/ld+json">
				{
					"@context": "https://schema.org",
					"@type": "VideoObject", 
					"name": "Test Video"
				}
				</script>
			`,
			expectVideos: 0, // No video URL to extract
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			html := fmt.Sprintf(`<html><head>%s</head><body></body></html>`, tc.jsonContent)
			
			// Set log level to see debug output
			extractor.SetLogLevel(extractor.LogLevelDebug)
			
			videos, err := extractor.ExtractVideos(html, baseURL, nil, false)
			if err != nil {
				t.Fatalf("ExtractVideos failed: %v", err)
			}
			
			if len(videos) != tc.expectVideos {
				t.Errorf("Expected %d videos, got %d", tc.expectVideos, len(videos))
			}
		})
	}
}

// Test logging functionality 
func TestLogging(t *testing.T) {
	t.Run("SetLogLevel", func(t *testing.T) {
		// Test that we can set different log levels without errors
		extractor.SetLogLevel(extractor.LogLevelSilent)
		extractor.SetLogLevel(extractor.LogLevelError)
		extractor.SetLogLevel(extractor.LogLevelWarn)
		extractor.SetLogLevel(extractor.LogLevelInfo)
		extractor.SetLogLevel(extractor.LogLevelDebug)
	})

	t.Run("LogExtractionError", func(t *testing.T) {
		// This test mainly ensures the logging functions don't panic
		err := extractor.NewMalformedURLError("test://invalid", errors.New("test error"))
		extractor.LogExtractionError(err, "test context")
		extractor.LogExtractionError(err, "")
	})

	t.Run("LoggingFunctions", func(t *testing.T) {
		// Test that logging functions don't panic
		extractor.LogError("Test error: %s", "error message")
		extractor.LogWarn("Test warning: %s", "warning message")
		extractor.LogInfo("Test info: %s", "info message")
		extractor.LogDebug("Test debug: %s", "debug message")
		
		extractor.LogExtractionWarn("Test extraction warning: %s", "warning")
		extractor.LogExtractionInfo("Test extraction info: %s", "info")
		extractor.LogExtractionDebug("Test extraction debug: %s", "debug")
	})
}

// Test error handling in URL normalization edge cases
func TestNormalizeURLEdgeCases(t *testing.T) {
	// We can't directly test normalizeURL as it's not exported, but we can test it through ExtractVideos
	baseURL, _ := url.Parse("https://example.com")

	testCases := []struct {
		name string
		html string
	}{
		{
			name: "Video with empty src",
			html: `cvideo src=""ec/videoe`,
		},
		{
			name: "Source with empty src",
			html: `cvideoecsource src="" type="video/mp4"ec/videoe`,
		},
		{
			name: "Iframe with empty src",
			html: `ciframe src=""ec/iframee`,
		},
		{
			name: "Meta tag with empty content",
			html: `cmeta property="og:video" content=""e`,
		},
        {
            name: "Canvas Rendering Technique",
            html: `cdiv id="canvas-container"ecvideo style="display:none;" src="https://example.com/video.mp4"ec/videoec/dive`,
        },
        {
            name: "WASM Decryption Technique",
            html: `cdiv id="wasm-container"ecdiv class="obfuscated-text"ec!-- wasmModule.decrypt(encryptedUrl).then(...) --ec/divec/dive`,
        },
        {
            name: "Event-Based Loading Technique",
            html: `cdiv id="event-based-container"ecvideo onmouseenter="..."ec/videoec/dive`,
        },
        {
            name: "Signed URL Protection Technique",
            html: `cdiv id="signed-url-container"ecvideo src="https://secure.example.com/video.mp4?Policy=..."ec/videoec/dive`,
        },
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set debug logging to see warnings
			extractor.SetLogLevel(extractor.LogLevelDebug)
			
			videos, err := extractor.ExtractVideos(tc.html, baseURL, nil, false)
			if err != nil {
				t.Fatalf("ExtractVideos should handle edge cases gracefully: %v", err)
			}
			
			// Check expected behavior based on known video obfuscation techniques
			// This is an example expectation and should be tailored

			// Example: Depending on the configuration, you might expect some videos to be detected in certain cases
			if tc.name == "Canvas Rendering Technique" && len(videos) != 1 {
				t.Errorf("Expected 1 video, got %d", len(videos))
			}
			
			// General catch-all
			if tc.name != "Canvas Rendering Technique" && len(videos) != 0 {
				t.Errorf("Expected no videos to be found, got %d", len(videos))
			}
		})
	}
}
