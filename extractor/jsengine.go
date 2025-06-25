package extractor

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/itxtx/crawler_go/config"
)

// UserInteraction holds parameters for simulating user interactions
type UserInteraction struct {
	Enabled       bool          `json:"enabled,omitempty"`
	MouseMovement bool          `json:"mouse_movement,omitempty"`
	MouseClicks   bool          `json:"mouse_clicks,omitempty"`
	ClickDelay    time.Duration `json:"click_delay,omitempty"`
	Scroll        bool          `json:"scroll,omitempty"`
	ScrollDelay   time.Duration `json:"scroll_delay,omitempty"`
}

// JSEngine provides JavaScript evaluation capabilities using headless Chrome
type JSEngine struct {
	ctx          context.Context
	cancel       context.CancelFunc
	initialized  bool
	timeout      time.Duration
	interaction  *UserInteraction
	tokenManager *TokenManager
}

// JSVideoResult represents a video found through JavaScript evaluation
type JSVideoResult struct {
	URL          string
	Title        string
	Thumbnail    string
	Platform     string
	Method       string // How it was discovered (js-eval, network-intercept, etc.)
	DynamicAttrs map[string]string
	IsExpiring   bool
	Token        string
}

// NewJSEngine creates a new JavaScript engine instance
func NewJSEngine(timeout time.Duration, interaction *UserInteraction, tokenManager *TokenManager) *JSEngine {
	return &JSEngine{
		timeout:      timeout,
		interaction:  interaction,
		tokenManager: tokenManager,
	}
}

// Initialize starts the headless browser
func (js *JSEngine) Initialize() error {
	if js.initialized {
		return nil
	}

	// Detect Chrome path from environment or system default
	chromePath := detectChromePath()
	if chromePath == "" {
		return fmt.Errorf("Chrome/Chromium not found. Please install Chrome/Chromium 118+ or set CHROME_PATH environment variable")
	}

	// Create Chrome options with custom path
	opts := []chromedp.ExecAllocatorOption{
		chromedp.ExecPath(chromePath),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-features", "VizDisplayCompositor"),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	}

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx)

	js.ctx = ctx
	js.cancel = cancel
	js.initialized = true

	LogExtractionInfo("JavaScript engine initialized with headless Chrome")
	return nil
}

// Close shuts down the JavaScript engine
func (js *JSEngine) Close() {
	if js.cancel != nil {
		js.cancel()
	}
	js.initialized = false
	LogExtractionInfo("JavaScript engine closed")
}

// ExtractVideosWithJS performs enhanced video extraction using JavaScript evaluation
func (js *JSEngine) ExtractVideosWithJS(targetURL string, baseURL *url.URL, cfg *config.CrawlerConfig) ([]VideoInfo, error) {
	if !js.initialized {
		if err := js.Initialize(); err != nil {
			return nil, fmt.Errorf("failed to initialize JS engine: %w", err)
		}
	}

	// Create a timeout context for this operation
	timeoutCtx, cancel := context.WithTimeout(js.ctx, js.timeout)
	defer cancel()

	var videos []VideoInfo
	var networkRequests []string
	var finalHTML string

	// Set up network request monitoring
	chromedp.ListenTarget(timeoutCtx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			reqURL := ev.Request.URL
			if isVideoURL(reqURL) {
				networkRequests = append(networkRequests, reqURL)
				LogExtractionDebug("Intercepted video network request: %s", reqURL)
			}
		}
	})

	// Run the page evaluation
	err := chromedp.Run(timeoutCtx,
		// Enable network monitoring
		network.Enable(),

		// Navigate to the page
		chromedp.Navigate(targetURL),

		// Wait for initial load
		chromedp.WaitReady("body", chromedp.ByQuery),

		// Simulate human-like behavior
		js.simulateHumanBehavior(),

		// Wait for dynamic content
		chromedp.Sleep(2*time.Second),

		// Execute JavaScript to find videos
		js.executeVideoDiscoveryJS(),

		// Trigger potential lazy loading
		js.triggerLazyLoading(),

		// Wait for additional content
		chromedp.Sleep(1*time.Second),

		// Get final HTML after JS execution
		chromedp.OuterHTML("html", &finalHTML),
	)

	if err != nil {
		LogExtractionError(NewJSEvaluationError(targetURL, err), "ExtractVideosWithJS")
		// Fallback to regular extraction if JS fails
		return ExtractVideos(finalHTML, baseURL, cfg, false)
	}

	// Process network-intercepted videos
	for _, reqURL := range networkRequests {
		absURL := normalizeURL(reqURL, baseURL)
		if absURL != "" {
			video := VideoInfo{
				URL:      absURL,
				Platform: inferPlatformFromURL(absURL),
				Title:    extractTitleFromURL(absURL),
			}
			videos = append(videos, video)
			LogExtractionDebug("Added network-intercepted video: %s", absURL)
		}
	}

	// Extract videos from the final HTML (after JS execution)
	htmlVideos, err := ExtractVideos(finalHTML, baseURL, cfg, false)
	if err != nil {
		LogExtractionWarn("Failed to extract videos from JS-processed HTML: %v", err)
	} else {
		videos = append(videos, htmlVideos...)
	}

	// Execute custom JavaScript patterns for obfuscated content
	jsVideos, err := js.executeCustomVideoExtraction(timeoutCtx, baseURL)
	if err != nil {
		LogExtractionWarn("Custom JS video extraction failed: %v", err)
	} else {
		videos = append(videos, jsVideos...)
	}

	// Monitor for dynamic content changes if enabled
	if cfg.HumanBehavior {
		dynamicMonitor := NewDynamicMonitor(true, 2*time.Second, 3)
		dynamicVideos, err := dynamicMonitor.MonitorContentChanges(timeoutCtx, js)
		if err != nil {
			LogExtractionWarn("Dynamic content monitoring failed: %v", err)
		} else {
			videos = append(videos, dynamicVideos...)
			LogExtractionInfo("Dynamic monitoring found %d additional videos", len(dynamicVideos))
		}
	}

	// Deduplicate videos
	return js.deduplicateVideos(videos), nil
}

// simulateHumanBehavior performs human-like interactions
func (js *JSEngine) simulateHumanBehavior() chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		// Scroll down to trigger lazy loading
		err := chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight/3)`, nil).Do(ctx)
		if err != nil {
			LogExtractionWarn("Failed to scroll during human behavior simulation: %v", err)
		}

		// Wait a bit
		time.Sleep(500 * time.Millisecond)

		// Move mouse to trigger hover events
		err = chromedp.MouseClickXY(100, 100).Do(ctx)
		if err != nil {
			LogExtractionWarn("Failed to simulate mouse click: %v", err)
		}

		return nil
	})
}

// executeVideoDiscoveryJS runs JavaScript to discover video elements
func (js *JSEngine) executeVideoDiscoveryJS() chromedp.Action {
	script := `
	(function() {
		const videos = [];
		
		// Find all video elements (including hidden ones)
		document.querySelectorAll('video').forEach(video => {
			const src = video.src || video.currentSrc;
			if (src) {
				videos.push({
					url: src,
					platform: 'html5',
					title: video.title || video.getAttribute('data-title') || '',
					poster: video.poster || ''
				});
			}
			
			// Check source elements
			video.querySelectorAll('source').forEach(source => {
				if (source.src) {
					videos.push({
						url: source.src,
						platform: 'html5',
						title: video.title || '',
						poster: video.poster || ''
					});
				}
			});
		});
		
		// Find data attributes that might contain video URLs
		document.querySelectorAll('[data-video-url], [data-src], [data-lazy-src]').forEach(el => {
			const url = el.getAttribute('data-video-url') || 
						el.getAttribute('data-src') || 
						el.getAttribute('data-lazy-src');
			if (url && url.includes('.mp4') || url.includes('.webm') || url.includes('.avi')) {
				videos.push({
					url: url,
					platform: 'custom',
					title: el.getAttribute('data-title') || el.title || ''
				});
			}
		});
		
		// Check for Video.js players
		if (window.videojs) {
			try {
				const players = videojs.getAllPlayers();
				players.forEach(player => {
					const src = player.currentSource();
					if (src && src.src) {
						videos.push({
							url: src.src,
							platform: 'videojs',
							title: player.el().getAttribute('data-title') || ''
						});
					}
				});
			} catch (e) {
				console.log('Error accessing Video.js players:', e);
			}
		}
		
		// Store results in a global variable for retrieval
		window._discoveredVideos = videos;
		return videos;
	})();
	`

	return chromedp.Evaluate(script, nil)
}

// triggerLazyLoading attempts to trigger lazy-loaded content
func (js *JSEngine) triggerLazyLoading() chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		// Scroll through the page to trigger lazy loading
		script := `
		(function() {
			const scrollHeight = document.body.scrollHeight;
			const steps = 5;
			let currentScroll = 0;
			
			function scrollStep() {
				currentScroll += scrollHeight / steps;
				window.scrollTo(0, currentScroll);
				
				// Trigger mouse events that might load videos
				const event = new MouseEvent('mouseover', {
					view: window,
					bubbles: true,
					cancelable: true
				});
				
				document.querySelectorAll('[data-video-url], [data-src], .video-container').forEach(el => {
					el.dispatchEvent(event);
				});
			}
			
			for (let i = 0; i < steps; i++) {
				setTimeout(scrollStep, i * 200);
			}
		})();
		`

		return chromedp.Evaluate(script, nil).Do(ctx)
	})
}

// executeCustomVideoExtraction performs custom JavaScript extraction patterns
func (js *JSEngine) executeCustomVideoExtraction(ctx context.Context, baseURL *url.URL) ([]VideoInfo, error) {
	var videos []VideoInfo

	// Get discovered videos from our JS
	var discoveredVideos []map[string]interface{}
	err := chromedp.Evaluate(`window._discoveredVideos || []`, &discoveredVideos).Do(ctx)
	if err != nil {
		return videos, err
	}

	for _, videoData := range discoveredVideos {
		url, ok := videoData["url"].(string)
		if !ok || url == "" {
			continue
		}

		absURL := normalizeURL(url, baseURL)
		if absURL == "" {
			continue
		}

		// Check and refresh token if URL token is expired
		if js.tokenManager != nil && js.tokenManager.IsTokenExpired(absURL) {
			newToken, err := js.tokenManager.RefreshToken(absURL)
			if err != nil {
				LogExtractionWarn("Failed to refresh token for %s: %v", absURL, err)
				continue
			}
			// Append new token to URL if it doesn't already have query params
			if strings.Contains(absURL, "?") {
				absURL = fmt.Sprintf("%s&token=%s", absURL, newToken)
			} else {
				absURL = fmt.Sprintf("%s?token=%s", absURL, newToken)
			}
			LogExtractionInfo("Token refreshed for video URL: %s", absURL)
		}

		platform, _ := videoData["platform"].(string)
		title, _ := videoData["title"].(string)
		poster, _ := videoData["poster"].(string)

		video := VideoInfo{
			URL:       absURL,
			Platform:  platform,
			Title:     title,
			Thumbnail: normalizeURL(poster, baseURL),
		}

		videos = append(videos, video)
		LogExtractionDebug("Added JS-discovered video: %s (platform: %s)", absURL, platform)
	}

	// Try to decode Base64 encoded video URLs
	base64Videos, err := js.decodeBase64VideoURLs(ctx, baseURL)
	if err != nil {
		LogExtractionWarn("Base64 video URL decoding failed: %v", err)
	} else {
		videos = append(videos, base64Videos...)
	}

	// Try to execute any WebAssembly modules that might decrypt video URLs
	wasmVideos, err := js.handleWebAssemblyDecryption(ctx, baseURL)
	if err != nil {
		LogExtractionWarn("WebAssembly video decryption failed: %v", err)
	} else {
		videos = append(videos, wasmVideos...)
	}

	return videos, nil
}

// decodeBase64VideoURLs attempts to decode Base64 encoded video URLs
func (js *JSEngine) decodeBase64VideoURLs(ctx context.Context, baseURL *url.URL) ([]VideoInfo, error) {
	script := `
	(function() {
		const videos = [];
		const base64Regex = /^[A-Za-z0-9+/]*={0,2}$/;
		
		// Look for data attributes that might be Base64 encoded
		document.querySelectorAll('[data-video], [data-encrypted], [data-encoded]').forEach(el => {
			const encoded = el.getAttribute('data-video') || 
						   el.getAttribute('data-encrypted') || 
						   el.getAttribute('data-encoded');
			
			if (encoded && encoded.length > 20 && base64Regex.test(encoded)) {
				try {
					const decoded = atob(encoded);
					// Check if decoded string looks like a URL
					if (decoded.startsWith('http') && (decoded.includes('.mp4') || decoded.includes('.webm'))) {
						videos.push({
							url: decoded,
							platform: 'base64-decoded',
							method: 'base64-decode'
						});
					}
				} catch (e) {
					// Not valid Base64 or not a video URL
				}
			}
		});
		
		return videos;
	})();
	`

	var decodedVideos []map[string]interface{}
	err := chromedp.Evaluate(script, &decodedVideos).Do(ctx)
	if err != nil {
		return nil, err
	}

	var videos []VideoInfo
	for _, videoData := range decodedVideos {
		url, ok := videoData["url"].(string)
		if !ok || url == "" {
			continue
		}

		absURL := normalizeURL(url, baseURL)
		if absURL == "" {
			continue
		}

		video := VideoInfo{
			URL:      absURL,
			Platform: "base64-decoded",
			Title:    extractTitleFromURL(absURL),
		}

		videos = append(videos, video)
		LogExtractionDebug("Decoded Base64 video URL: %s", absURL)
	}

	return videos, nil
}

// handleWebAssemblyDecryption attempts to handle WebAssembly-based video URL decryption
func (js *JSEngine) handleWebAssemblyDecryption(ctx context.Context, baseURL *url.URL) ([]VideoInfo, error) {
	script := `
	(function() {
		const videos = [];
		
		// Check if there are any WASM modules loaded
		if (typeof WebAssembly !== 'undefined') {
			// Look for global functions that might decrypt video URLs
			const globalKeys = Object.keys(window);
			const decryptFunctions = globalKeys.filter(key => 
				key.toLowerCase().includes('decrypt') || 
				key.toLowerCase().includes('decode') ||
				key.toLowerCase().includes('video')
			);
			
			decryptFunctions.forEach(funcName => {
				try {
					const func = window[funcName];
					if (typeof func === 'function') {
						// Try to call with common encrypted patterns
						const testPatterns = [
							document.querySelector('[data-encrypted]')?.getAttribute('data-encrypted'),
							document.querySelector('[data-video-encrypted]')?.getAttribute('data-video-encrypted')
						].filter(Boolean);
						
						testPatterns.forEach(pattern => {
							try {
								const result = func(pattern);
								if (typeof result === 'string' && result.startsWith('http') && 
								   (result.includes('.mp4') || result.includes('.webm'))) {
									videos.push({
										url: result,
										platform: 'wasm-decrypted',
										method: 'wasm-decrypt',
										decryptFunction: funcName
									});
								}
							} catch (e) {
								// Function call failed, continue
							}
						});
					}
				} catch (e) {
					// Error accessing function, continue
				}
			});
		}
		
		return videos;
	})();
	`

	var wasmVideos []map[string]interface{}
	err := chromedp.Evaluate(script, &wasmVideos).Do(ctx)
	if err != nil {
		return nil, err
	}

	var videos []VideoInfo
	for _, videoData := range wasmVideos {
		url, ok := videoData["url"].(string)
		if !ok || url == "" {
			continue
		}

		absURL := normalizeURL(url, baseURL)
		if absURL == "" {
			continue
		}

		video := VideoInfo{
			URL:      absURL,
			Platform: "wasm-decrypted",
			Title:    extractTitleFromURL(absURL),
		}

		videos = append(videos, video)
		LogExtractionDebug("WASM-decrypted video URL: %s", absURL)
	}

	return videos, nil
}

// deduplicateVideos removes duplicate video entries
func (js *JSEngine) deduplicateVideos(videos []VideoInfo) []VideoInfo {
	seen := make(map[string]bool)
	var unique []VideoInfo

	for _, video := range videos {
		if !seen[video.URL] {
			seen[video.URL] = true
			unique = append(unique, video)
		}
	}

	return unique
}

// detectChromePath detects Chrome installation path from environment or system defaults
func detectChromePath() string {
	// Check CHROME_PATH environment variable first
	if chromePath := os.Getenv("CHROME_PATH"); chromePath != "" {
		// Verify the path exists
		if _, err := os.Stat(chromePath); err == nil {
			return chromePath
		}
		LogExtractionWarn("CHROME_PATH is set but file does not exist: %s", chromePath)
	}

	// Platform-specific default paths
	var candidates []string
	switch runtime.GOOS {
	case "darwin": // macOS
		candidates = []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
		}
	case "linux":
		candidates = []string{
			"/usr/bin/google-chrome",
			"/usr/bin/google-chrome-stable",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
			"/snap/bin/chromium",
			"/opt/google/chrome/chrome",
		}
	case "windows":
		candidates = []string{
			"C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
			"C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe",
			"C:\\Users\\" + os.Getenv("USERNAME") + "\\AppData\\Local\\Google\\Chrome\\Application\\chrome.exe",
			// WSL paths
			"/mnt/c/Program Files/Google/Chrome/Application/chrome.exe",
			"/mnt/c/Program Files (x86)/Google/Chrome/Application/chrome.exe",
		}
	default:
		// For other systems, try common Linux paths
		candidates = []string{
			"/usr/bin/google-chrome",
			"/usr/bin/chromium",
		}
	}

	// Check each candidate path
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			LogExtractionDebug("Found Chrome at: %s", path)
			return path
		}
	}

	// If no Chrome found, log available paths for debugging
	LogExtractionWarn("Chrome not found in any of the following locations:")
	for _, path := range candidates {
		LogExtractionWarn("  - %s", path)
	}
	LogExtractionWarn("Please install Chrome/Chromium or set CHROME_PATH environment variable")

	return ""
}

// ExtractVideosWithJavaScript is a convenience function that creates a JS engine and extracts videos
func ExtractVideosWithJavaScript(targetURL string, baseURL *url.URL, cfg *config.CrawlerConfig, timeout time.Duration) ([]VideoInfo, error) {
	// Create default interaction and token manager
	interaction := &UserInteraction{
		Enabled:       cfg.HumanBehavior,
		MouseMovement: true,
		MouseClicks:   true,
		Scroll:        true,
		ClickDelay:    500 * time.Millisecond,
		ScrollDelay:   500 * time.Millisecond,
	}
	
	tokenManager := NewTokenManager("", "") // Default empty token manager
	
	engine := NewJSEngine(timeout, interaction, tokenManager)
	defer engine.Close()

	return engine.ExtractVideosWithJS(targetURL, baseURL, cfg)
}
