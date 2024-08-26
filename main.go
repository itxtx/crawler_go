package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

type config struct {
	pages              map[string]int
	baseURL            *url.URL
	mu                 *sync.Mutex
	concurrencyControl chan struct{}
	wg                 *sync.WaitGroup
	maxPage            int
}

func (cfg *config) addPageVisit(normalizedURL string) bool {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	if _, exists := cfg.pages[normalizedURL]; exists {
		cfg.pages[normalizedURL]++
		return false
	}

	cfg.pages[normalizedURL] = 1
	return true
}

func (cfg *config) crawlPage(rawCurrentURL string) {
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

		// Check if we've already crawled this page
		if !cfg.addPageVisit(normalizedURL) {
			return
		}

		// Get the HTML from the current URL
		fmt.Printf("Crawling: %s\n", normalizedURL)
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

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go <base_url>")
		os.Exit(1)
	}

	baseURL, err := url.Parse(os.Args[1])
	if err != nil {
		fmt.Println("Error parsing base URL:", err)
		os.Exit(1)
	}

	fmt.Printf("Starting crawl of: %s\n", baseURL)

	cfg := &config{
		pages:              make(map[string]int),
		baseURL:            baseURL,
		mu:                 &sync.Mutex{},
		concurrencyControl: make(chan struct{}, 25), // Adjust this value to control concurrency
		wg:                 &sync.WaitGroup{},
		maxPage:            100, // Adjust this value to control the maximum number of pages to crawl
	}

	cfg.crawlPage(baseURL.String())

	cfg.wg.Wait()

	fmt.Println("\nCrawl results:")
	for url, count := range cfg.pages {
		fmt.Printf("%s: %d\n", url, count)
	}
}
