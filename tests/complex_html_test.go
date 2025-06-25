package main

import (
	"net/url"
	"strings"
	"testing"

	"github.com/itxtx/crawler_go/config"
	"github.com/itxtx/crawler_go/extractor"
)

// TestComplexVideoObfuscationHTML tests the crawler against a complex HTML document
// with advanced video protection techniques
func TestComplexVideoObfuscationHTML(t *testing.T) {
	baseURL, _ := url.Parse("https://example.com")
	
	// The complete HTML content from the provided document
	complexHTML := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Advanced Video Obfuscation Techniques</title>
    <link href="https://vjs.zencdn.net/8.10.0/video-js.css" rel="stylesheet">
    <style>
        /* Extensive CSS styles omitted for brevity */
        .video-js { width: 100% !important; height: 100% !important; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Advanced Video Protection System</h1>
        <p class="subtitle">Dynamic embedding techniques with multiple layers of obfuscation</p>
    </div>
    
    <div class="container">
        <div class="section">
            <h2 class="section-title">🔒 Enhanced Video Protection Techniques</h2>
            
            <div class="technique-grid">
                <div class="technique-card">
                    <h3 class="card-title">🎨 Canvas Rendering</h3>
                    <div class="card-content">
                        <p>The video element is hidden; its frames are drawn onto a visible canvas, hiding the source.</p>
                        <div class="video-container">
                            <div class="protection-level">Level 6 Security</div>
                            <div class="loading-indicator">
                                <div class="spinner"></div>
                                <div>Initializing canvas stream...</div>
                            </div>
                            <div class="video-wrapper" id="canvas-container"></div>
                            <div class="video-overlay"></div>
                        </div>
                        <div class="obfuscated-text">
                            // const video = document.createElement('video');
                            // video.style.display = 'none';
                            // const canvasCtx = canvas.getContext('2d');
                            // requestAnimationFrame(drawFrame);
                        </div>
                    </div>
                    <div class="complexity high">Complexity: High</div>
                </div>
                
                <div class="technique-card">
                    <h3 class="card-title">⚙️ WebAssembly Decryption</h3>
                    <div class="card-content">
                        <p>A secure, hard-to-reverse-engineer WebAssembly module decrypts the video URL at runtime.</p>
                        <div class="video-container">
                            <div class="protection-level">Level 7 Security</div>
                            <div class="loading-indicator">
                                <div class="spinner"></div>
                                <div>Awaiting WASM decryption...</div>
                            </div>
                            <div class="video-wrapper" id="wasm-container"></div>
                            <div class="video-overlay"></div>
                        </div>
                        <div class="obfuscated-text" id="wasm-code">
                            // Encrypted data from server
                            const encryptedUrl = "aHR0cHM6Ly9...==";
                            // wasmModule.decrypt(encryptedUrl).then(...)
                        </div>
                    </div>
                    <div class="complexity high">Complexity: High</div>
                </div>

                <div class="technique-card">
                    <h3 class="card-title">🖱️ Event-Based Loading</h3>
                    <div class="card-content">
                        <p>Video is only loaded after a specific user interaction, like moving the mouse over the area.</p>
                        <div class="video-container">
                            <div class="protection-level">Level 4 Security</div>
                            <div class="loading-indicator">
                                 <div class="event-prompt">▶ Move mouse here to load</div>
                            </div>
                            <div class="video-wrapper" id="event-based-container"></div>
                            <div class="video-overlay"></div>
                        </div>
                        <div class="obfuscated-text">
                            // container.addEventListener('mouseenter', () => {
                            //     // ... create video element ...
                            // }, { once: true });
                        </div>
                    </div>
                    <div class="complexity medium">Complexity: Medium</div>
                </div>
                
                <div class="technique-card">
                    <h3 class="card-title">🔐 Signed URL Protection</h3>
                    <div class="card-content">
                        <p>Video sources use time-limited signed URLs with complex query parameters.</p>
                        <div class="video-container">
                            <div class="protection-level">Level 5 Security</div>
                            <div class="loading-indicator">
                                <div class="spinner"></div>
                                <div>Verifying access token...</div>
                            </div>
                            <div class="video-wrapper" id="signed-url-container"></div>
                            <div class="video-overlay"></div>
                        </div>
                        <div class="obfuscated-text">
                            &lt;video src="https://secure.example.com/video.mp4?Policy=ey...&amp;Signature=...&amp;Key-Pair-Id=..."&gt;
                        </div>
                    </div>
                    <div class="complexity high">Complexity: High</div>
                </div>
            </div>
        </div>
    </div>
    
    <div class="footer">
        <p>Advanced Video Protection System | Obfuscation Techniques for Secure Embedding</p>
        <p>Note: This is a demonstration of security techniques</p>
    </div>

    <script src="https://vjs.zencdn.net/8.10.0/video.min.js"></script>
    <script>
        document.addEventListener('DOMContentLoaded', function() {
            // JavaScript implementation for dynamic video loading...
            // This would normally contain the actual obfuscation logic
            console.log('Video protection system initialized');
        });
    </script>
</body>
</html>`

	// Set debug logging to see what's happening
	extractor.SetLogLevel(extractor.LogLevelDebug)

	t.Run("Basic extraction from complex HTML", func(t *testing.T) {
		videos, err := extractor.ExtractVideos(complexHTML, baseURL, nil, false)
		if err != nil {
			t.Fatalf("ExtractVideos should handle complex HTML gracefully: %v", err)
		}

		// This HTML doesn't contain actual video elements in the traditional sense,
		// but may contain URLs in text that match video patterns
		t.Logf("Found %d videos in complex obfuscated HTML", len(videos))

		// The HTML contains a reference to "https://secure.example.com/video.mp4?Policy=ey...&Signature=...&Key-Pair-Id=..."
		// which should be detected by the direct URL extraction
		hasSignedURL := false
		for _, video := range videos {
			if strings.Contains(video.URL, "secure.example.com") {
				hasSignedURL = true
				if video.Platform != "direct" {
					t.Errorf("Expected platform 'direct' for signed URL, got '%s'", video.Platform)
				}
			}
		}

		if !hasSignedURL {
			t.Error("Expected to find the signed URL in the obfuscated text")
		}
	})

	t.Run("Custom selector for container IDs", func(t *testing.T) {
		// Test CSS selectors targeting the specific container IDs
		cfg := &config.CrawlerConfig{
			Selectors:    []string{"#canvas-container", "#wasm-container", "#event-based-container", "#signed-url-container"},
			SelectorType: "css",
		}

		videos, err := extractor.ExtractVideos(complexHTML, baseURL, cfg, false)
		if err != nil {
			t.Fatalf("ExtractVideos should handle custom selectors gracefully: %v", err)
		}

		t.Logf("Found %d videos with container ID selectors", len(videos))
		// These containers don't have actual video sources, so we might not find any
	})

	t.Run("Custom regex for obfuscated URLs", func(t *testing.T) {
		// Test regex to find URLs in obfuscated text
		cfg := &config.CrawlerConfig{
			Selectors:    []string{`https://[^\s"'&]+\.mp4[^\s"']*`},
			SelectorType: "regex",
		}

		videos, err := extractor.ExtractVideos(complexHTML, baseURL, cfg, false)
		if err != nil {
			t.Fatalf("ExtractVideos should handle regex selectors gracefully: %v", err)
		}

		t.Logf("Found %d videos with regex selectors", len(videos))
		
		// Should find the signed URL pattern
		foundSecureURL := false
		for _, video := range videos {
			if strings.Contains(video.URL, "secure.example.com") {
				foundSecureURL = true
				if video.Platform != "custom" {
					t.Errorf("Expected platform 'custom' for regex-found URL, got '%s'", video.Platform)
				}
			}
		}

		if !foundSecureURL {
			t.Error("Regex should find the secure URL pattern")
		}
	})

	t.Run("XPath selector for obfuscated text", func(t *testing.T) {
		// Test XPath to find elements with obfuscated text class
		cfg := &config.CrawlerConfig{
			Selectors:    []string{"//div[@class='obfuscated-text']"},
			SelectorType: "xpath",
		}

		videos, err := extractor.ExtractVideos(complexHTML, baseURL, cfg, false)
		if err != nil {
			t.Fatalf("ExtractVideos should handle XPath selectors gracefully: %v", err)
		}

		t.Logf("Found %d videos with XPath selectors", len(videos))
		// These divs don't have src attributes, so we might not find any videos
	})

	t.Run("Error handling with malformed obfuscated content", func(t *testing.T) {
		// Test with partially malformed HTML
		malformedHTML := complexHTML[:len(complexHTML)/2] + `<div class="malformed"`

		videos, err := extractor.ExtractVideos(malformedHTML, baseURL, nil, false)
		if err != nil {
			t.Fatalf("ExtractVideos should handle malformed complex HTML gracefully: %v", err)
		}

		t.Logf("Found %d videos in malformed complex HTML", len(videos))
	})

	t.Run("Performance with large complex HTML", func(t *testing.T) {
		// Test with repeated complex HTML to simulate a large document
		largeHTML := strings.Repeat(complexHTML, 5)

		videos, err := extractor.ExtractVideos(largeHTML, baseURL, nil, false)
		if err != nil {
			t.Fatalf("ExtractVideos should handle large HTML gracefully: %v", err)
		}

		t.Logf("Found %d videos in large HTML document", len(videos))
		
		// Should handle deduplication properly
		urlSet := make(map[string]bool)
		for _, video := range videos {
			if urlSet[video.URL] {
				t.Errorf("Found duplicate video URL: %s", video.URL)
			}
			urlSet[video.URL] = true
		}
	})
}

// TestVideoObfuscationTechniques tests specific video obfuscation scenarios
func TestVideoObfuscationTechniques(t *testing.T) {
	baseURL, _ := url.Parse("https://example.com")

	testCases := []struct {
		name          string
		html          string
		expectVideos  int
		expectPlatform string
		description   string
	}{
		{
			name: "Canvas rendering with hidden video",
			html: `
				<div id="canvas-container">
					<video style="display:none;" src="https://example.com/hidden.mp4"></video>
					<canvas id="display-canvas"></canvas>
				</div>
			`,
			expectVideos:  1,
			expectPlatform: "html5",
			description:   "Hidden video should still be detected",
		},
		{
			name: "JavaScript-generated video URLs",
			html: `
				<script>
					const videoUrl = "https://example.com/dynamic.mp4";
					// More JavaScript obfuscation
				</script>
				<div>Video URL in script: https://example.com/script-video.mp4</div>
			`,
			expectVideos:  1,
			expectPlatform: "direct",
			description:   "URLs in text should be detected",
		},
		{
			name: "Base64 encoded video references",
			html: `
				<div data-video="aHR0cHM6Ly9leGFtcGxlLmNvbS92aWRlby5tcDQ=">
					<!-- Base64: https://example.com/video.mp4 -->
				</div>
			`,
			expectVideos:  0, // Base64 decoding not implemented in basic version
			expectPlatform: "",
			description:   "Base64 encoded URLs not detected without decoding",
		},
		{
			name: "Video.js player initialization",
			html: `
				<video class="video-js" data-setup='{}' controls>
					<source src="https://example.com/videojs.mp4" type="video/mp4">
				</video>
				<script src="https://vjs.zencdn.net/8.10.0/video.min.js"></script>
			`,
			expectVideos:  1,
			expectPlatform: "html5",
			description:   "Video.js sources should be detected",
		},
		{
			name: "Multiple protection layers",
			html: `
				<div class="protection-level">Level 7 Security</div>
				<div class="obfuscated-text">
					const encrypted = "encrypted_video_url";
					// https://secure.example.com/protected.mp4?token=abc123
				</div>
				<iframe src="https://www.youtube.com/embed/dQw4w9WgXcQ" style="display:none;"></iframe>
			`,
			expectVideos:  2, // Direct URL + YouTube iframe
			expectPlatform: "direct", // First found would be direct
			description:   "Multiple sources with different protection levels",
		},
		{
			name: "Event-driven video loading",
			html: `
				<div id="video-placeholder" 
					 data-src="https://example.com/lazy.mp4"
					 onclick="loadVideo(this)">
					Click to load video
				</div>
				<script>
					function loadVideo(element) {
						// Dynamic video creation
					}
				</script>
			`,
			expectVideos:  1,
			expectPlatform: "custom",
			description:   "Data attributes should be detected",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			extractor.SetLogLevel(extractor.LogLevelDebug)
			
			videos, err := extractor.ExtractVideos(tc.html, baseURL, nil, false)
			if err != nil {
				t.Fatalf("ExtractVideos failed for %s: %v", tc.description, err)
			}

			if len(videos) != tc.expectVideos {
				t.Errorf("Expected %d videos for %s, got %d", tc.expectVideos, tc.description, len(videos))
				for i, video := range videos {
					t.Logf("  Video %d: %s (platform: %s)", i+1, video.URL, video.Platform)
				}
			}

			if tc.expectVideos > 0 && len(videos) > 0 {
				found := false
				for _, video := range videos {
					if video.Platform == tc.expectPlatform {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find platform '%s' for %s", tc.expectPlatform, tc.description)
				}
			}
		})
	}
}

// TestRobustnessAgainstObfuscation tests the crawler's robustness against various obfuscation techniques
func TestRobustnessAgainstObfuscation(t *testing.T) {
	baseURL, _ := url.Parse("https://example.com")

	t.Run("Heavily nested and obfuscated structure", func(t *testing.T) {
		obfuscatedHTML := `
			<div class="layer1">
				<div class="layer2" style="position:relative;">
					<div class="layer3" data-protection="enabled">
						<div class="video-wrapper" id="secure-container">
							<div class="overlay-protection"></div>
							<div class="content-mask">
								<video src="https://example.com/deeply-nested.mp4" 
									   style="opacity:0;position:absolute;top:-9999px;">
								</video>
							</div>
							<canvas class="display-surface"></canvas>
						</div>
					</div>
				</div>
			</div>
		`

		videos, err := extractor.ExtractVideos(obfuscatedHTML, baseURL, nil, false)
		if err != nil {
			t.Fatalf("Should handle heavily nested structure: %v", err)
		}

		if len(videos) != 1 {
			t.Errorf("Expected 1 video despite heavy nesting, got %d", len(videos))
		}
	})

	t.Run("Mixed content with distractors", func(t *testing.T) {
		mixedHTML := `
			<!-- Fake video references -->
			<div class="fake-video">fake://not-a-video.mp4</div>
			<div>Some text about video.mp4 files in general</div>
			
			<!-- Real video hidden among distractors -->
			<section class="media-section">
				<p>This section contains media files including:</p>
				<ul>
					<li>Audio files: music.mp3, sound.wav</li>
					<li>Video files: intro.mp4, main.avi</li>
					<li>Images: banner.jpg, thumbnail.png</li>
				</ul>
				<video controls>
					<source src="https://example.com/real-video.mp4" type="video/mp4">
					<source src="https://example.com/real-video.webm" type="video/webm">
				</video>
			</section>
			
			<!-- More distractors -->
			<div>Download video.mp4 files from our server</div>
		`

		videos, err := extractor.ExtractVideos(mixedHTML, baseURL, nil, false)
		if err != nil {
			t.Fatalf("Should handle mixed content: %v", err)
		}

		// Should find 2 real videos (mp4 and webm sources)
		realVideoCount := 0
		for _, video := range videos {
			if strings.Contains(video.URL, "real-video") {
				realVideoCount++
			}
		}

		if realVideoCount != 2 {
			t.Errorf("Expected 2 real videos, found %d", realVideoCount)
		}
	})
}
