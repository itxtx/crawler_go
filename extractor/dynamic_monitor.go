package extractor

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
)

// DynamicMonitor handles periodic re-evaluation of pages for content changes
type DynamicMonitor struct {
	enabled    bool
	interval   time.Duration
	maxRetries int
	videos     []VideoInfo
}

// NewDynamicMonitor creates a new dynamic content monitor
func NewDynamicMonitor(enabled bool, interval time.Duration, maxRetries int) *DynamicMonitor {
	return &DynamicMonitor{
		enabled:    enabled,
		interval:   interval,
		maxRetries: maxRetries,
		videos:     make([]VideoInfo, 0),
	}
}

// MonitorContentChanges periodically re-evaluates the page for new video content
func (dm *DynamicMonitor) MonitorContentChanges(ctx context.Context, js *JSEngine) ([]VideoInfo, error) {
	if !dm.enabled {
		return dm.videos, nil
	}

	// Monitor for changes over the specified interval
	ticker := time.NewTicker(dm.interval)
	defer ticker.Stop()

	retries := 0
	for retries < dm.maxRetries {
		select {
		case <-ctx.Done():
			return dm.videos, ctx.Err()
		case <-ticker.C:
			// Re-evaluate the page for new video content
			newVideos, err := dm.evaluateForNewContent(ctx)
			if err != nil {
				LogExtractionWarn("Failed to evaluate for new content: %v", err)
				retries++
				continue
			}

			// Add any new videos found
			for _, video := range newVideos {
				if !dm.videoExists(video) {
					dm.videos = append(dm.videos, video)
					LogExtractionInfo("Detected new dynamic video: %s", video.URL)
				}
			}
			
			retries++
		}
	}

	return dm.videos, nil
}

// evaluateForNewContent executes JavaScript to check for new video content
func (dm *DynamicMonitor) evaluateForNewContent(ctx context.Context) ([]VideoInfo, error) {
	script := `
	(function() {
		const videos = [];
		
		// Check for newly loaded video elements
		document.querySelectorAll('video:not([data-monitored])').forEach(video => {
			const src = video.src || video.currentSrc;
			if (src) {
				videos.push({
					url: src,
					platform: 'html5-dynamic',
					title: video.title || video.getAttribute('data-title') || '',
					poster: video.poster || ''
				});
				video.setAttribute('data-monitored', 'true');
			}
		});
		
		// Check for AJAX-loaded content
		const ajaxElements = document.querySelectorAll('[data-video-url]:not([data-monitored])');
		ajaxElements.forEach(el => {
			const url = el.getAttribute('data-video-url');
			if (url) {
				videos.push({
					url: url,
					platform: 'ajax-loaded',
					title: el.getAttribute('data-title') || el.title || ''
				});
				el.setAttribute('data-monitored', 'true');
			}
		});
		
		return videos;
	})();
	`

	var dynamicVideos []map[string]interface{}
	err := chromedp.Evaluate(script, &dynamicVideos).Do(ctx)
	if err != nil {
		return nil, err
	}

	var videos []VideoInfo
	for _, videoData := range dynamicVideos {
		url, ok := videoData["url"].(string)
		if !ok || url == "" {
			continue
		}

		platform, _ := videoData["platform"].(string)
		title, _ := videoData["title"].(string)
		poster, _ := videoData["poster"].(string)

		video := VideoInfo{
			URL:       url,
			Platform:  platform,
			Title:     title,
			Thumbnail: poster,
		}

		videos = append(videos, video)
	}

	return videos, nil
}

// videoExists checks if a video already exists in the monitored list
func (dm *DynamicMonitor) videoExists(newVideo VideoInfo) bool {
	for _, existingVideo := range dm.videos {
		if existingVideo.URL == newVideo.URL {
			return true
		}
	}
	return false
}
