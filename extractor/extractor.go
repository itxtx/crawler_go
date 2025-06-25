package extractor

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xpath"
	"github.com/itxtx/crawler_go/config"
	"golang.org/x/net/html"
)

type LinkInfo struct {
	URL         string
	Description string
}

func formatOutput(matches []string, format string) string {
	switch format {
	case "json":
		jsonData, _ := json.Marshal(matches)
		return string(jsonData)
	case "csv":
		return strings.Join(matches, ",")
	default:
		return strings.Join(matches, "\n")
	}
}

func ExtractContent(htmlContent, pattern, selectorType, format string, printContent bool) {
	var matches []string
	switch selectorType {
	case "css":
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
		if err != nil {
			fmt.Println("Error parsing HTML:", err)
			return
		}
		doc.Find(pattern).Each(func(i int, s *goquery.Selection) {
			matches = append(matches, s.Text())
		})
	case "xpath":
		doc, err := htmlquery.Parse(strings.NewReader(htmlContent))
		if err != nil {
			fmt.Println("Error parsing HTML:", err)
			return
		}
		expr := xpath.MustCompile(pattern)
		nodes := htmlquery.Find(doc, expr.String())
		for _, node := range nodes {
			matches = append(matches, htmlquery.InnerText(node))
		}
	case "regex":
		re := regexp.MustCompile(pattern)
		reMatches := re.FindAllStringSubmatch(htmlContent, -1)
		for _, match := range reMatches {
			if len(match) > 1 {
				matches = append(matches, match[1])
			}
		}
	}

	if len(matches) == 0 {
		fmt.Println("No content matched the pattern.")
		return
	}

	output := formatOutput(matches, format)
	fmt.Println(output)

	fmt.Println("Extracted content:")
	for _, match := range matches {
		fmt.Println(match)
	}
}

func ExtractMultipleContents(htmlContent string, patterns []string, selectorType string, format string, printContent bool) {
	for _, pattern := range patterns {
		ExtractContent(htmlContent, pattern, selectorType, format, printContent)
	}
}

func ExtractLinksAndDescriptions(htmlBody string, baseURL *url.URL, filter string) ([]LinkInfo, error) {
	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		return nil, err
	}

	var links []LinkInfo
	var extractFunc func(*html.Node)
	extractFunc = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			var href, text string
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					href = attr.Val
					break
				}
			}
			if href != "" {
				absoluteURL := baseURL.ResolveReference(&url.URL{Path: href}).String()
				if filter == "" || strings.Contains(absoluteURL, filter) {
					text = extractText(n)
					links = append(links, LinkInfo{URL: absoluteURL, Description: text})
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractFunc(c)
		}
	}
	extractFunc(doc)
	return links, nil
}

func extractText(n *html.Node) string {
	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			text += c.Data
		} else if c.Type == html.ElementNode {
			text += extractText(c)
		}
	}
	return strings.TrimSpace(text)
}

// ExtractVideos extracts video information from HTML content using multiple detection strategies
func ExtractVideos(htmlContent string, base *url.URL, cfg *config.CrawlerConfig, printContent bool) ([]VideoInfo, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %w", err)
	}

	// Use a map for deduplication
	videoMap := make(map[string]VideoInfo)

	// Strategy 1: HTML5 video <video> and nested <source> tags
	extractHTML5Videos(doc, base, videoMap)

	// Strategy 2: <iframe> embeds for YouTube, Vimeo, etc.
	extractIframeEmbeds(doc, base, videoMap)

	// Handle custom selectors when selector_type targets video (run early for priority)
	if cfg != nil && len(cfg.Selectors) > 0 && (strings.Contains(strings.ToLower(cfg.SelectorType), "video") || cfg.SelectorType == "css" || cfg.SelectorType == "xpath" || cfg.SelectorType == "regex") {
		for _, selector := range cfg.Selectors {
			switch cfg.SelectorType {
			case "css":
				doc.Find(selector).Each(func(i int, s *goquery.Selection) {
					if src, exists := s.Attr("src"); exists {
						addVideoToMap(src, "custom", base, s, doc, videoMap)
					}
					// Also check for href attributes for links to videos
					if href, exists := s.Attr("href"); exists {
						if isVideoURL(href) {
							addVideoToMap(href, "custom", base, s, doc, videoMap)
						}
					}
					// Also check data attributes
					if dataURL, exists := s.Attr("data-video-url"); exists {
						addVideoToMap(dataURL, "custom", base, s, doc, videoMap)
					}
				})
			case "xpath":
				if htmlDoc, err := htmlquery.Parse(strings.NewReader(htmlContent)); err == nil {
					// Validate XPath expression
					if _, err := xpath.Compile(selector); err != nil {
						LogExtractionError(NewXPathParseError(selector, err), "ExtractVideos")
						continue
					}

					nodes := htmlquery.Find(htmlDoc, selector)
					for _, node := range nodes {
						if src := htmlquery.SelectAttr(node, "src"); src != "" {
							addVideoToMap(src, "custom", base, nil, doc, videoMap)
						}
						if href := htmlquery.SelectAttr(node, "href"); href != "" && isVideoURL(href) {
							addVideoToMap(href, "custom", base, nil, doc, videoMap)
						}
					}
				} else {
					LogExtractionError(NewHTMLParseError(htmlContent, err), "ExtractVideos XPath")
				}
			case "regex":
				re, err := regexp.Compile(selector)
				if err != nil {
					LogExtractionError(NewRegexParseError(selector, err), "ExtractVideos")
					continue
				}
				reMatches := re.FindAllStringSubmatch(htmlContent, -1)
				for _, match := range reMatches {
					if len(match) > 1 {
						addVideoToMap(match[1], "custom", base, nil, doc, videoMap)
					}
				}
			}
		}
	}

	// Strategy 3: Direct video URLs in text content
	extractDirectVideoURLs(htmlContent, base, videoMap)

	// Strategy 4: JSON-LD schema.org/VideoObject
	extractJSONLDVideos(htmlContent, base, videoMap)

	// Strategy 5: Open Graph / Twitter meta tags
	extractMetaTagVideos(doc, base, videoMap)

	// Strategy 6: Custom data attributes
	extractCustomDataAttributes(doc, base, videoMap)

	// Convert map to slice
	videos := make([]VideoInfo, 0, len(videoMap))
	for _, video := range videoMap {
		videos = append(videos, video)
	}

	// Print content if requested, similar to ExtractContent pattern
	if printContent && len(videos) > 0 {
		format := "text" // default format
		if cfg != nil && cfg.VideoFormat != "" {
			format = cfg.VideoFormat
		}
		output := FormatVideoOutput(videos, format)
		fmt.Println(output)
	}

	return videos, nil
}

// extractHTML5Videos extracts videos from HTML5 <video> and <source> tags
func extractHTML5Videos(doc *goquery.Document, base *url.URL, videoMap map[string]VideoInfo) {
	doc.Find("video").Each(func(i int, video *goquery.Selection) {
		// Check for src attribute on video tag
		if src, exists := video.Attr("src"); exists {
			addVideoToMap(src, "html5", base, video, doc, videoMap)
		}

		// Check for source tags within video
		video.Find("source").Each(func(j int, source *goquery.Selection) {
			if src, exists := source.Attr("src"); exists {
				addVideoToMap(src, "html5", base, video, doc, videoMap)
			}
		})

		// Also check for poster attribute for thumbnail
		if poster, exists := video.Attr("poster"); exists {
			if src, srcExists := video.Attr("src"); srcExists {
				absURL := normalizeURL(src, base)
				if existing, found := videoMap[absURL]; found {
					existing.Thumbnail = normalizeURL(poster, base)
					videoMap[absURL] = existing
				}
			}
		}
	})
}

// extractIframeEmbeds extracts videos from iframe embeds (YouTube, Vimeo, etc.)
func extractIframeEmbeds(doc *goquery.Document, base *url.URL, videoMap map[string]VideoInfo) {
	doc.Find("iframe").Each(func(i int, iframe *goquery.Selection) {
		if src, exists := iframe.Attr("src"); exists {
			platform := inferPlatformFromURL(src)
			if platform != "unknown" {
				addVideoToMap(src, platform, base, iframe, doc, videoMap)
			}
		}
	})
}

// extractDirectVideoURLs extracts direct video URLs from HTML content using regex
func extractDirectVideoURLs(htmlContent string, base *url.URL, videoMap map[string]VideoInfo) {
	// Regex for direct video URLs, but be more selective to avoid false positives
	videoURLRegex := regexp.MustCompile(`(?i)https?://[^"'\s<>]+\.(mp4|webm|avi|mov|mkv|flv|wmv|m4v)(?:[?#][^\s"'<>]*)?`)
	matches := videoURLRegex.FindAllString(htmlContent, -1)

	for _, match := range matches {
		// Skip URLs that are clearly Base64 encoded data or other non-direct URLs
		if isLikelyEncodedOrObfuscated(match) {
			LogExtractionDebug("Skipping potentially encoded/obfuscated URL: %s", match)
			continue
		}
		
		absURL := normalizeURL(match, base)
		if absURL != "" {
			if _, exists := videoMap[absURL]; !exists {
				videoMap[absURL] = VideoInfo{
					URL:      absURL,
					Platform: "direct",
					Title:    extractTitleFromURL(absURL),
				}
				LogExtractionDebug("Found direct video URL: %s", absURL)
			}
		}
	}
}

// extractJSONLDVideos extracts videos from JSON-LD schema.org/VideoObject
func extractJSONLDVideos(htmlContent string, base *url.URL, videoMap map[string]VideoInfo) {
	if htmlContent == "" {
		LogExtractionWarn("Empty HTML content provided for JSON-LD extraction")
		return
	}

	// Find JSON-LD script tags
	jsonLDRegex := regexp.MustCompile(`(?s)\<script[^\>]*type=["']application/ld\+json["'][^\>]*\>(.*?)\</script\>`)
	matches := jsonLDRegex.FindAllStringSubmatch(htmlContent, -1)

	for _, match := range matches {
		if len(match) > 1 {
			jsonContent := strings.TrimSpace(match[1])
			if jsonContent == "" {
				LogExtractionWarn("Empty JSON-LD content found")
				continue
			}

			var data interface{}
			if err := json.Unmarshal([]byte(jsonContent), &data); err != nil {
				LogExtractionError(NewJSONLDParseError(jsonContent, err), "extractJSONLDVideos")
				continue
			}
			extractVideoFromJSONLD(data, base, videoMap)
		}
	}
}

// extractVideoFromJSONLD recursively extracts video objects from JSON-LD data
func extractVideoFromJSONLD(data interface{}, base *url.URL, videoMap map[string]VideoInfo) {
	switch v := data.(type) {
	case map[string]interface{}:
		if typeVal, ok := v["@type"].(string); ok && typeVal == "VideoObject" {
			var video VideoInfo
			video.Platform = "schema"

			if url, ok := v["contentUrl"].(string); ok {
				video.URL = normalizeURL(url, base)
			} else if url, ok := v["url"].(string); ok {
				video.URL = normalizeURL(url, base)
			}

			if title, ok := v["name"].(string); ok {
				video.Title = title
			}

			if thumbnail, ok := v["thumbnailUrl"].(string); ok {
				video.Thumbnail = normalizeURL(thumbnail, base)
			}

			if width, ok := v["width"].(float64); ok {
				video.Width = int(width)
			}

			if height, ok := v["height"].(float64); ok {
				video.Height = int(height)
			}

			if video.URL != "" {
				videoMap[video.URL] = video
			}
		}

		// Recursively check nested objects
		for _, val := range v {
			extractVideoFromJSONLD(val, base, videoMap)
		}

	case []interface{}:
		for _, item := range v {
			extractVideoFromJSONLD(item, base, videoMap)
		}
	}
}

// extractMetaTagVideos extracts videos from Open Graph and Twitter meta tags
func extractMetaTagVideos(doc *goquery.Document, base *url.URL, videoMap map[string]VideoInfo) {
	// Open Graph video tags
	ogVideoSelectors := []string{
		"meta[property='og:video']",
		"meta[property='og:video:url']",
		"meta[property='og:video:secure_url']",
	}

	for _, selector := range ogVideoSelectors {
		doc.Find(selector).Each(func(i int, meta *goquery.Selection) {
			if content, exists := meta.Attr("content"); exists {
				addMetaVideo(content, "og", base, doc, videoMap)
			}
		})
	}

	// Twitter player tags
	doc.Find("meta[name='twitter:player']").Each(func(i int, meta *goquery.Selection) {
		if content, exists := meta.Attr("content"); exists {
			addMetaVideo(content, "twitter", base, doc, videoMap)
		}
	})
}

// extractCustomDataAttributes extracts videos from custom data attributes
func extractCustomDataAttributes(doc *goquery.Document, base *url.URL, videoMap map[string]VideoInfo) {
	dataAttributes := []string{
		"data-video-src",
		"data-video-url",
		"data-src",
		"data-video",
		"data-lazy",
		"data-lazy-src",
		"data-source",
	}

	// Check standard data attributes
	for _, attr := range dataAttributes {
		doc.Find(fmt.Sprintf("[%s]", attr)).Each(func(i int, element *goquery.Selection) {
			if src, exists := element.Attr(attr); exists && src != "" {
				// Skip Base64 encoded data URIs unless they decode to valid URLs
				if !isLikelyEncodedOrObfuscated(src) && isVideoURL(src) {
					addVideoToMap(src, "custom", base, element, doc, videoMap)
				}
			}
		})
	}

	// Also check for any element with onclick handlers that might load videos dynamically
	doc.Find("[onclick]").Each(func(i int, element *goquery.Selection) {
		// Look for data-src or similar attributes on clickable elements
		for _, attr := range []string{"data-src", "data-video-url", "data-lazy"} {
			if src, exists := element.Attr(attr); exists && src != "" {
				if !isLikelyEncodedOrObfuscated(src) && isVideoURL(src) {
					addVideoToMap(src, "custom", base, element, doc, videoMap)
				}
			}
		}
	})
}

// addVideoToMap adds a video to the map with metadata extraction
func addVideoToMap(src, platform string, base *url.URL, element *goquery.Selection, doc *goquery.Document, videoMap map[string]VideoInfo) {
	if src == "" {
		LogExtractionWarn("Empty src provided to addVideoToMap")
		return
	}

	absURL := normalizeURL(src, base)
	if absURL == "" {
		// Error already logged by normalizeURL
		return
	}

	if _, exists := videoMap[absURL]; exists {
		LogExtractionDebug("Video already exists in map: %s", absURL)
		return // Already exists
	}

	video := VideoInfo{
		URL:      absURL,
		Platform: platform,
	}

	// Try to extract title
	video.Title = extractVideoTitle(element, doc)

	// Try to extract thumbnail
	video.Thumbnail = extractVideoThumbnail(element, doc, base)

	// Try to extract dimensions
	video.Width, video.Height = extractVideoDimensions(element)

	videoMap[absURL] = video
	LogExtractionDebug("Added video to map: %s (platform: %s)", absURL, platform)
}

// addMetaVideo adds a video from meta tags
func addMetaVideo(content, platform string, base *url.URL, doc *goquery.Document, videoMap map[string]VideoInfo) {
	absURL := normalizeURL(content, base)
	if absURL == "" {
		return
	}

	if _, exists := videoMap[absURL]; exists {
		return
	}

	video := VideoInfo{
		URL:      absURL,
		Platform: platform,
	}

	// Extract title from og:title or twitter:title
	if title := doc.Find("meta[property='og:title']").AttrOr("content", ""); title != "" {
		video.Title = title
	} else if title := doc.Find("meta[name='twitter:title']").AttrOr("content", ""); title != "" {
		video.Title = title
	} else if title := doc.Find("title").Text(); title != "" {
		video.Title = title
	}

	// Extract thumbnail from og:image or twitter:image
	if thumb := doc.Find("meta[property='og:image']").AttrOr("content", ""); thumb != "" {
		video.Thumbnail = normalizeURL(thumb, base)
	} else if thumb := doc.Find("meta[name='twitter:image']").AttrOr("content", ""); thumb != "" {
		video.Thumbnail = normalizeURL(thumb, base)
	}

	videoMap[absURL] = video
}

// normalizeURL converts relative URLs to absolute URLs with error handling
func normalizeURL(rawURL string, base *url.URL) string {
	if rawURL == "" {
		LogExtractionWarn("Empty URL provided for normalization")
		return ""
	}

	if base == nil {
		LogExtractionError(NewMalformedURLError(rawURL, fmt.Errorf("no base URL provided")), "normalizeURL")
		return ""
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		LogExtractionError(NewMalformedURLError(rawURL, err), "normalizeURL")
		return ""
	}

	return base.ResolveReference(parsed).String()
}

// inferPlatformFromURL infers the video platform from URL patterns
func inferPlatformFromURL(rawURL string) string {
	lowerURL := strings.ToLower(rawURL)

	switch {
	case strings.Contains(lowerURL, "youtube.com") || strings.Contains(lowerURL, "youtu.be"):
		return "youtube"
	case strings.Contains(lowerURL, "vimeo.com"):
		return "vimeo"
	case strings.Contains(lowerURL, "dailymotion.com"):
		return "dailymotion"
	case strings.Contains(lowerURL, "twitch.tv"):
		return "twitch"
	case strings.Contains(lowerURL, "facebook.com"):
		return "facebook"
	case strings.Contains(lowerURL, "instagram.com"):
		return "instagram"
	case strings.Contains(lowerURL, "tiktok.com"):
		return "tiktok"
	default:
		return "unknown"
	}
}

// extractVideoTitle tries to extract a title for the video
func extractVideoTitle(element *goquery.Selection, doc *goquery.Document) string {
	// Handle nil element case
	if element == nil {
		if title := doc.Find("title").Text(); title != "" {
			return title
		}
		return ""
	}
	
	// Try title attribute first
	if title, exists := element.Attr("title"); exists && title != "" {
		return title
	}

	// Try alt attribute
	if alt, exists := element.Attr("alt"); exists && alt != "" {
		return alt
	}

	// Try aria-label
	if label, exists := element.Attr("aria-label"); exists && label != "" {
		return label
	}

	// Try nearby text content
	if text := element.Siblings().First().Text(); text != "" {
		return strings.TrimSpace(text)
	}

	// Try parent text content
	if text := element.Parent().Text(); text != "" {
		return strings.TrimSpace(text)
	}

	// Fallback to page title
	if title := doc.Find("title").Text(); title != "" {
		return title
	}

	return ""
}

// extractVideoThumbnail tries to extract a thumbnail for the video
func extractVideoThumbnail(element *goquery.Selection, doc *goquery.Document, base *url.URL) string {
	// Handle nil element case
	if element == nil {
		return ""
	}
	
	// Try poster attribute (for video tags)
	if poster, exists := element.Attr("poster"); exists {
		return normalizeURL(poster, base)
	}

	// Try data-poster or similar attributes
	posterAttrs := []string{"data-poster", "data-thumbnail", "data-thumb"}
	for _, attr := range posterAttrs {
		if poster, exists := element.Attr(attr); exists {
			return normalizeURL(poster, base)
		}
	}

	// Try to find nearby img tags
	var thumbnail string
	element.Siblings().Find("img").Each(func(i int, img *goquery.Selection) {
		if src, exists := img.Attr("src"); exists && thumbnail == "" {
			thumbnail = normalizeURL(src, base)
		}
	})

	if thumbnail != "" {
		return thumbnail
	}

	// Try parent's img tags
	element.Parent().Find("img").Each(func(i int, img *goquery.Selection) {
		if src, exists := img.Attr("src"); exists && thumbnail == "" {
			thumbnail = normalizeURL(src, base)
		}
	})

	return thumbnail
}

// extractVideoDimensions tries to extract video dimensions
func extractVideoDimensions(element *goquery.Selection) (int, int) {
	width := 0
	height := 0

	// Handle nil element case
	if element == nil {
		return width, height
	}

	if w, exists := element.Attr("width"); exists {
		if parsed, err := strconv.Atoi(w); err == nil {
			width = parsed
		}
	}

	if h, exists := element.Attr("height"); exists {
		if parsed, err := strconv.Atoi(h); err == nil {
			height = parsed
		}
	}

	return width, height
}

// extractTitleFromURL extracts a basic title from a URL
func extractTitleFromURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	// Get the filename without extension
	path := parsed.Path
	if lastSlash := strings.LastIndex(path, "/"); lastSlash != -1 {
		path = path[lastSlash+1:]
	}

	if lastDot := strings.LastIndex(path, "."); lastDot != -1 {
		path = path[:lastDot]
	}

	// Replace underscores and hyphens with spaces
	path = strings.ReplaceAll(path, "_", " ")
	path = strings.ReplaceAll(path, "-", " ")

	return strings.TrimSpace(path)
}

// isVideoURL checks if a URL points to a video file or video platform
func isVideoURL(rawURL string) bool {
	lowerURL := strings.ToLower(rawURL)
	
	// Check for video file extensions
	videoExtensions := []string{".mp4", ".webm", ".avi", ".mov", ".mkv", ".flv", ".wmv", ".m4v"}
	for _, ext := range videoExtensions {
		if strings.Contains(lowerURL, ext) {
			return true
		}
	}
	
	// Check for video platforms
	videoHosts := []string{"youtube.com", "youtu.be", "vimeo.com", "dailymotion.com", "twitch.tv", "facebook.com/watch", "instagram.com/p", "tiktok.com"}
	for _, host := range videoHosts {
		if strings.Contains(lowerURL, host) {
			return true
		}
	}
	
	return false
}

// isLikelyEncodedOrObfuscated checks if a URL appears to be encoded or obfuscated
func isLikelyEncodedOrObfuscated(rawURL string) bool {
	// Check for Base64-like patterns (long strings of alphanumeric characters with = padding)
	if len(rawURL) > 50 && strings.Contains(rawURL, "=") {
		// Count alphanumeric characters vs total length
		alphanumCount := 0
		for _, char := range rawURL {
			if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '+' || char == '/' || char == '=' {
				alphanumCount++
			}
		}
		// If more than 80% of characters are Base64-like, it's likely encoded
		if float64(alphanumCount)/float64(len(rawURL)) > 0.8 {
			return true
		}
	}
	
	// Check for extremely long path segments (potential obfuscation)
	if parsed, err := url.Parse(rawURL); err == nil {
		pathSegments := strings.Split(parsed.Path, "/")
		for _, segment := range pathSegments {
			if len(segment) > 100 { // Very long path segment
				return true
			}
		}
	}
	
	// Check for data URIs
	if strings.HasPrefix(rawURL, "data:") {
		return true
	}
	
	// Check for blob URLs
	if strings.HasPrefix(rawURL, "blob:") {
		return true
	}
	
	return false
}
