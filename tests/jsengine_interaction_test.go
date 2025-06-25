package main

import (
	"testing"
	"time"

	"github.com/itxtx/crawler_go/extractor"
)

func TestUserInteraction(t *testing.T) {
	tm := extractor.NewTokenManager("https://api.example.com/refresh", "dummy_api_key")
	interaction := &extractor.UserInteraction{
		Enabled:       true,
		MouseMovement: true,
		MouseClicks:   true,
		Scroll:        true,
		ClickDelay:    500 * time.Millisecond,
		ScrollDelay:   500 * time.Millisecond,
	}

	js := extractor.NewJSEngine(30*time.Second, interaction, tm)
	ts := js.Initialize()
	if ts != nil {
		t.Errorf("failed to initialize Javascript engine: %v", ts)
	}
	
	js.Close()
}

func TestTokenManagement(t *testing.T) {
	// Test with empty endpoint (should fallback to mock behavior)
	tm := extractor.NewTokenManager("", "dummy_api_key")

	videoURL := "https://example.com/video.mp4"
	
	// Test IsTokenExpired (should always return false in current implementation)
	isExpired := tm.IsTokenExpired(videoURL)
	if isExpired {
		t.Logf("Token is expired for video: %s", videoURL)
	} else {
		t.Logf("Token is not expired for video: %s", videoURL)
	}

	// Test RefreshToken with empty endpoint (should return error)
	refreshedToken, err := tm.RefreshToken(videoURL)
	if err != nil {
		t.Logf("Expected error with empty endpoint: %v", err)
	} else {
		t.Errorf("Expected error with empty endpoint, but got token: %s", refreshedToken)
	}

	// Test with non-empty endpoint (but we expect it to fail)
	tm2 := extractor.NewTokenManager("https://mock.api.com/refresh", "test_key")
	_, err2 := tm2.RefreshToken(videoURL)
	if err2 != nil {
		t.Logf("Expected network error with mock endpoint: %v", err2)
	}

	t.Log("Token management test completed successfully")
}

func TestDynamicContentMonitoring(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping dynamic content monitoring test in short mode")
	}

	// Test dynamic monitor initialization
	dynamicMonitor := extractor.NewDynamicMonitor(true, 1*time.Second, 2)
	if dynamicMonitor == nil {
		t.Error("Failed to create dynamic monitor")
	}

	// Test with disabled monitor
	disabledMonitor := extractor.NewDynamicMonitor(false, 1*time.Second, 2)
	if disabledMonitor == nil {
		t.Error("Failed to create disabled dynamic monitor")
	}

	t.Log("Dynamic content monitoring test completed successfully")
}
