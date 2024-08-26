package main

import (
	"fmt"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)

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

func main() {
	// Ensure there are at least 4 arguments (program name + 3 required args)
	if len(os.Args) < 4 {
		fmt.Println("Usage: ./crawler <base_url> <max_concurrency> <max_pages> [selectors] [selector_type] [output_format] [filter]")
		os.Exit(1)
	}

	// Parse required arguments
	baseURL, err := url.Parse(os.Args[1])
	if err != nil {
		fmt.Println("Error parsing base URL:", err)
		os.Exit(1)
	}

	maxConcurrency, err := strconv.Atoi(os.Args[2])
	if err != nil || maxConcurrency < 1 {
		fmt.Println("Invalid max_concurrency. Must be a positive integer.")
		os.Exit(1)
	}

	maxPages, err := strconv.Atoi(os.Args[3])
	if err != nil || maxPages < 1 {
		fmt.Println("Invalid max_pages. Must be a positive integer.")
		os.Exit(1)
	}

	// Parse optional arguments with default values
	var selectors []string
	selectorType := "css" // default selector type
	format := "text"      // default output format
	filter := ""          // default no filter

	if len(os.Args) > 4 { // e.g., "h1", "p", "a"
		selectors = strings.Split(os.Args[4], ",")
	}
	if len(os.Args) > 5 { //e.g., "css", "xpath", "regex"
		selectorType = os.Args[5]
	}
	if len(os.Args) > 6 { // e.g., "json", "csv", "text"
		format = os.Args[6]
	}
	if len(os.Args) > 7 { // e.g., "Blog Post"
		filter = os.Args[7]
	}

	fmt.Printf("Starting crawl of: %s\n", baseURL)
	fmt.Printf("Max concurrency: %d\n", maxConcurrency)
	fmt.Printf("Max pages: %d\n", maxPages)
	fmt.Printf("Selectors: %v\n", selectors)
	fmt.Printf("Selector Type: %s\n", selectorType)
	fmt.Printf("Output Format: %s\n", format)
	fmt.Printf("Filter: %s\n", filter)

	cfg := &config{
		pages:              make(map[string]int),
		baseURL:            baseURL,
		mu:                 &sync.Mutex{},
		concurrencyControl: make(chan struct{}, maxConcurrency),
		wg:                 &sync.WaitGroup{},
		maxPages:           maxPages,
	}

	// Start crawling with filtering if a filter is provided
	cfg.crawlPage(baseURL.String(), filter)

	cfg.wg.Wait()

	// Example of extracting content with optional selectors and type
	if len(selectors) > 0 {
		// Assuming htmlBody is retrieved within the crawling process
		htmlBody := "" // replace with actual HTML body content
		extractMultipleContents(htmlBody, selectors, selectorType, format, true)
		extractContent(htmlBody, strings.Join(selectors, ","), selectorType, format, true)
	}

	printReport(cfg.pages, baseURL.String())
}
