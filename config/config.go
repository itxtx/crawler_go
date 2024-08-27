package config

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
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

func ParseArgs(args []string) (*CrawlerConfig, error) {
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

func PrintReport(pages map[string]int, baseURL string) {
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
