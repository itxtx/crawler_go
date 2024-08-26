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

type config struct {
	pages              map[string]int
	baseURL            *url.URL
	mu                 *sync.Mutex
	concurrencyControl chan struct{}
	wg                 *sync.WaitGroup
	maxPages           int
}

func (cfg *config) addPageVisit(normalizedURL string) bool {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	if len(cfg.pages) >= cfg.maxPages {
		return false
	}

	if _, exists := cfg.pages[normalizedURL]; exists {
		cfg.pages[normalizedURL]++
		return false
	}

	cfg.pages[normalizedURL] = 1
	return true
}

func (cfg *config) crawlPage(rawCurrentURL string) {
	cfg.mu.Lock()
	if len(cfg.pages) >= cfg.maxPages {
		cfg.mu.Unlock()
		return
	}
	cfg.mu.Unlock()

	cfg.wg.Add(1)
	go func() {
		defer cfg.wg.Done()
		cfg.concurrencyControl <- struct{}{}
		defer func() { <-cfg.concurrencyControl }()

		currentURL, err := url.Parse(rawCurrentURL)
		if err != nil {
			fmt.Println("Error parsing current URL:", err)
			return
		}

		// Make sure the current URL is on the same domain as the base URL
		if cfg.baseURL.Hostname() != currentURL.Hostname() {
			return
		}

		// Normalize the current URL
		normalizedURL, err := normalizeURL(currentURL.String())
		if err != nil {
			fmt.Println("Error normalizing URL:", err)
			return
		}

		// Check if we've already crawled this page or reached the max pages
		if !cfg.addPageVisit(normalizedURL) {
			return
		}

		// Get the HTML from the current URL
		//fmt.Printf("Crawling: %s\n", normalizedURL)
		htmlBody, err := getHTML(currentURL.String())
		if err != nil {
			fmt.Println("Error fetching URL:", err)
			return
		}

		// Extract URLs from the HTML
		urls, err := getURLsFromHTML(htmlBody, currentURL.String())
		if err != nil {
			fmt.Println("Error extracting URLs:", err)
			return
		}

		// Recursively crawl each URL on the page
		for _, u := range urls {
			cfg.crawlPage(u)
		}
	}()
}

func getHTML(rawURL string) (string, error) {
	resp, err := http.Get(rawURL)
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
	if len(os.Args) != 4 {
		fmt.Println("Usage: ./crawler <base_url> <max_concurrency> <max_pages>")
		os.Exit(1)
	}

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

	fmt.Printf("Starting crawl of: %s\n", baseURL)
	fmt.Printf("Max concurrency: %d\n", maxConcurrency)
	fmt.Printf("Max pages: %d\n", maxPages)

	cfg := &config{
		pages:              make(map[string]int),
		baseURL:            baseURL,
		mu:                 &sync.Mutex{},
		concurrencyControl: make(chan struct{}, maxConcurrency),
		wg:                 &sync.WaitGroup{},
		maxPages:           maxPages,
	}

	cfg.crawlPage(baseURL.String())

	cfg.wg.Wait()

	printReport(cfg.pages, baseURL.String())

}
