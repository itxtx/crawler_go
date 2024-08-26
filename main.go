package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type CrawlerConfig struct {
	BaseURL        *url.URL
	MaxConcurrency int
	MaxPages       int
	Selectors      []string
	SelectorType   string
	OutputFormat   string
	Filter         string
}

type LinkInfo struct {
	URL         string
	Description string
}

type config struct {
	links              []LinkInfo
	pages              map[string]int
	baseURL            *url.URL
	mu                 *sync.Mutex
	concurrencyControl chan struct{}
	wg                 *sync.WaitGroup
	maxPages           int
	crawlerConfig      *CrawlerConfig
}

func fetchContent(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("received error status code: %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/html") {
		return "", fmt.Errorf("content is not text/html")
	}

	htmlData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(htmlData), nil
}

func parseArgs(args []string) (*CrawlerConfig, error) {
	if len(args) < 4 {
		return nil, fmt.Errorf("usage: ./crawler <base_url> <max_concurrency> <max_pages> [selectors] [selector_type] [output_format] [filter]")
	}

	baseURL, err := url.Parse(args[1])
	if err != nil {
		return nil, fmt.Errorf("error parsing base URL: %v", err)
	}

	maxConcurrency, err := strconv.Atoi(args[2])
	if err != nil || maxConcurrency < 1 {
		return nil, fmt.Errorf("invalid max_concurrency. Must be a positive integer")
	}

	maxPages, err := strconv.Atoi(args[3])
	if err != nil || maxPages < 1 {
		return nil, fmt.Errorf("invalid max_pages. Must be a positive integer")
	}

	config := &CrawlerConfig{
		BaseURL:        baseURL,
		MaxConcurrency: maxConcurrency,
		MaxPages:       maxPages,
		Selectors:      nil,
		SelectorType:   "css",  // default
		OutputFormat:   "text", // default
		Filter:         "",     // no filter by default
	}

	for i := 4; i < len(args); i++ {
		switch {
		case strings.Contains(args[i], "selectors="):
			config.Selectors = strings.Split(strings.TrimPrefix(args[i], "selectors="), ",")
		case strings.Contains(args[i], "selector_type="):
			config.SelectorType = strings.TrimPrefix(args[i], "selector_type=")
		case strings.Contains(args[i], "output_format="):
			config.OutputFormat = strings.TrimPrefix(args[i], "output_format=")
		case strings.Contains(args[i], "filter="):
			config.Filter = strings.TrimPrefix(args[i], "filter=")
		}
	}

	return config, nil
}

func printReport(pages map[string]int, baseURL string) {
	fmt.Println("=============================")
	fmt.Println("REPORT for", baseURL)
	fmt.Println("=============================")

	// Extract keys from the map and sort them
	var urls []string
	for url := range pages {
		urls = append(urls, url)
	}
	sort.Strings(urls)

	// Print the report with sorted URLs
	for _, url := range urls {
		fmt.Printf("Found %d internal links to %s\n", pages[url], url)
	}
}

func (cfg *config) addLink(link LinkInfo) bool {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	if len(cfg.links) >= cfg.maxPages {
		return false
	}

	cfg.links = append(cfg.links, link)
	return true
}

func main() {
	configArgs, err := parseArgs(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Starting crawl of: %s\n", configArgs.BaseURL)
	fmt.Printf("Max concurrency: %d\n", configArgs.MaxConcurrency)
	fmt.Printf("Max pages: %d\n", configArgs.MaxPages)
	fmt.Printf("Selectors: %v\n", configArgs.Selectors)
	fmt.Printf("Selector Type: %s\n", configArgs.SelectorType)
	fmt.Printf("Output Format: %s\n", configArgs.OutputFormat)
	fmt.Printf("Filter: %s\n", configArgs.Filter)

	cfg := &config{
		links:              make([]LinkInfo, 0, configArgs.MaxPages),
		pages:              make(map[string]int),
		baseURL:            configArgs.BaseURL,
		mu:                 &sync.Mutex{},
		concurrencyControl: make(chan struct{}, configArgs.MaxConcurrency),
		wg:                 &sync.WaitGroup{},
		maxPages:           configArgs.MaxPages,
		crawlerConfig:      configArgs,
	}

	// Start crawling with filtering if a filter is provided
	cfg.crawlPage(configArgs.BaseURL.String(), configArgs.Filter)

	cfg.wg.Wait()

	fmt.Println("\nCrawl results:")
	for _, link := range cfg.links {
		fmt.Printf("URL: %s\nDescription: %s\n\n", link.URL, link.Description)
	}

	printReport(cfg.pages, configArgs.BaseURL.String())
}
