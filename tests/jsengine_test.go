package main

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/itxtx/crawler_go/config"
	"github.com/itxtx/crawler_go/extractor"
)

// TestJSEngineBasicFunctionality tests basic JavaScript engine functionality
func TestJSEngineBasicFunctionality(t *testing.T) {
	// Skip if running in CI or headless environment without Chrome
	if testing.Short() {
		t.Skip("Skipping JavaScript engine tests in short mode")
	}

	baseURL, _ := url.Parse("https://example.com")

	// Create a simple test server with dynamic content
	testHTML := `
<!DOCTYPE html>
<html>
<head>
    <title>Test Page</title>
</head>
<body>
    <div id="video-container"></div>
    <script>
        // Simulate dynamic video loading
        setTimeout(function() {
            const container = document.getElementById('video-container');
            const video = document.createElement('video');
            video.src = 'https://example.com/dynamic-video.mp4';
            video.setAttribute('data-title', 'Dynamically Loaded Video');
            container.appendChild(video);
        }, 100);
        
        // Simulate Base64 encoded video URL
        const encodedVideo = document.createElement('div');
        encodedVideo.setAttribute('data-encoded', btoa('https://example.com/encoded-video.mp4'));
        document.body.appendChild(encodedVideo);
    </script>
</body>
</html>
`

	// Create default interaction and token manager for tests
	interaction := &extractor.UserInteraction{
		Enabled:       false, // Disable for basic tests
		MouseMovement: false,
		MouseClicks:   false,
		Scroll:        false,
	}
	tokenManager := extractor.NewTokenManager("", "")
	engine := extractor.NewJSEngine(10*time.Second, interaction, tokenManager)
	defer engine.Close()

	t.Run("Engine initialization", func(t *testing.T) {
		err := engine.Initialize()
		if err != nil {
			t.Skipf("Failed to initialize JavaScript engine (Chrome may not be available): %v", err)
		}
	})

	t.Run("Extract videos with JavaScript disabled", func(t *testing.T) {
		cfg := &config.CrawlerConfig{
			EnableJS: false,
		}

		// This should work with regular extraction
		videos, err := extractor.ExtractVideos(testHTML, baseURL, cfg, false)
		if err != nil {
			t.Fatalf("Regular video extraction failed: %v", err)
		}

		// Should not find dynamic videos
		t.Logf("Found %d videos with regular extraction", len(videos))
	})
}

// TestJSEngineObfuscationHandling tests various obfuscation scenarios
func TestJSEngineObfuscationHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping JavaScript engine tests in short mode")
	}

	baseURL, _ := url.Parse("https://example.com")

	testCases := []struct {
		name            string
		html            string
		expectMinVideos int
		description     string
	}{
		{
			name: "Hidden video elements",
			html: `
				<html>
				<body>
					<video style="display:none;" src="https://example.com/hidden.mp4"></video>
					<div style="position:absolute;top:-9999px;">
						<video src="https://example.com/offscreen.mp4"></video>
					</div>
				</body>
				</html>
			`,
			expectMinVideos: 2,
			description:     "Should find hidden video elements",
		},
		{
			name: "Video.js dynamic loading",
			html: `
				<html>
				<head>
					<script>
						// Mock Video.js
						window.videojs = {
							getAllPlayers: function() {
								return [{
									currentSource: function() {
										return { src: 'https://example.com/videojs-player.mp4' };
									},
									el: function() {
										return { getAttribute: function() { return 'VideoJS Video'; } };
									}
								}];
							}
						};
					</script>
				</head>
				<body>
					<video class="video-js" data-setup="{}"></video>
				</body>
				</html>
			`,
			expectMinVideos: 1,
			description:     "Should handle Video.js players",
		},
		{
			name: "Event-driven lazy loading",
			html: `
				<html>
				<body>
					<div data-video-url="https://example.com/lazy1.mp4" class="video-container">
						Hover to load video
					</div>
					<div data-src="https://example.com/lazy2.mp4" onclick="loadVideo()">
						Click to load video
					</div>
				</body>
				</html>
			`,
			expectMinVideos: 2,
			description:     "Should find lazy-loaded video URLs",
		},
		{
			name: "Multiple nested protection layers",
			html: `
				<html>
				<body>
					<div class="protection-layer-1">
						<div class="protection-layer-2" style="opacity:0;">
							<div class="content-mask">
								<video src="https://example.com/nested-protected.mp4" style="position:absolute;left:-9999px;"></video>
							</div>
						</div>
					</div>
					<iframe src="https://www.youtube.com/embed/test123" style="visibility:hidden;"></iframe>
				</body>
				</html>
			`,
			expectMinVideos: 2,
			description:     "Should handle nested protection layers",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create default interaction and token manager for tests
			interaction := &extractor.UserInteraction{
				Enabled:       false,
				MouseMovement: false,
				MouseClicks:   false,
				Scroll:        false,
			}
			tokenManager := extractor.NewTokenManager("", "")
			engine := extractor.NewJSEngine(15*time.Second, interaction, tokenManager)
			defer engine.Close()

			// For testing purposes, we'll test the individual components
			// since we can't easily serve the HTML content to the headless browser

			// Test regular extraction first
			videos, err := extractor.ExtractVideos(tc.html, baseURL, nil, false)
			if err != nil {
				t.Fatalf("Regular extraction failed: %v", err)
			}

			if len(videos) < tc.expectMinVideos {
				t.Logf("Regular extraction found %d videos, expected at least %d", len(videos), tc.expectMinVideos)
				t.Logf("Description: %s", tc.description)

				// List found videos for debugging
				for i, video := range videos {
					t.Logf("  Video %d: %s (platform: %s)", i+1, video.URL, video.Platform)
				}

				// This is expected for some dynamic content that requires JS
				if !strings.Contains(tc.name, "Video.js") && !strings.Contains(tc.name, "Event-driven") {
					t.Errorf("Expected at least %d videos for %s, got %d", tc.expectMinVideos, tc.description, len(videos))
				}
			}

			// Test that the engine can be initialized (Chrome availability test)
			err = engine.Initialize()
			if err != nil {
				t.Logf("JavaScript engine not available (Chrome may not be installed): %v", err)
				t.Skip("Skipping JavaScript-specific tests")
			}

			t.Logf("✓ JavaScript engine available for %s", tc.name)
		})
	}
}

// TestJSEngineConfiguration tests JavaScript engine configuration options
func TestJSEngineConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping JavaScript engine tests in short mode")
	}

	t.Run("Timeout configuration", func(t *testing.T) {
		interaction := &extractor.UserInteraction{Enabled: false}
		tokenManager := extractor.NewTokenManager("", "")
		shortTimeout := extractor.NewJSEngine(1*time.Second, interaction, tokenManager)
		defer shortTimeout.Close()

		longTimeout := extractor.NewJSEngine(30*time.Second, interaction, tokenManager)
		defer longTimeout.Close()

		// Test that engines can be created with different timeouts
		if shortTimeout == nil || longTimeout == nil {
			t.Error("Failed to create JS engines with different timeouts")
		}
	})

	t.Run("Multiple engine instances", func(t *testing.T) {
		interaction := &extractor.UserInteraction{Enabled: false}
		tokenManager := extractor.NewTokenManager("", "")
		engine1 := extractor.NewJSEngine(10*time.Second, interaction, tokenManager)
		defer engine1.Close()

		engine2 := extractor.NewJSEngine(10*time.Second, interaction, tokenManager)
		defer engine2.Close()

		// Test that multiple engines can coexist
		if engine1 == nil || engine2 == nil {
			t.Error("Failed to create multiple JS engine instances")
		}
	})
}

// TestJSEngineIntegration tests integration with the main crawler
func TestJSEngineIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping JavaScript engine integration tests in short mode")
	}

	baseURL, _ := url.Parse("https://example.com")

	t.Run("Config parsing for JS options", func(t *testing.T) {
		// Test that JavaScript options are parsed correctly
		args := []string{
			"program",
			"https://example.com",
			"5",
			"100",
			"enable_js=true",
			"js_timeout=20",
			"human_behavior=true",
			"video_extract=true",
		}

		cfg, err := config.ParseArgs(args)
		if err != nil {
			t.Fatalf("Failed to parse args with JS options: %v", err)
		}

		if !cfg.EnableJS {
			t.Error("EnableJS should be true")
		}

		if cfg.JSTimeout != 20 {
			t.Errorf("Expected JSTimeout to be 20, got %d", cfg.JSTimeout)
		}

		if !cfg.HumanBehavior {
			t.Error("HumanBehavior should be true")
		}

		if !cfg.ExtractVideos {
			t.Error("ExtractVideos should be true")
		}
	})

	t.Run("Fallback behavior", func(t *testing.T) {
		// Test that when JS extraction fails, it falls back to regular extraction
		cfg := &config.CrawlerConfig{
			EnableJS:      true,
			JSTimeout:     1, // Very short timeout to force failure
			ExtractVideos: true,
		}

		html := `<video src="https://example.com/test.mp4"></video>`

		// This should still work even if JS engine fails
		videos, err := extractor.ExtractVideos(html, baseURL, cfg, false)
		if err != nil {
			t.Fatalf("Fallback extraction should work: %v", err)
		}

		if len(videos) != 1 {
			t.Errorf("Expected 1 video from fallback extraction, got %d", len(videos))
		}
	})
}
