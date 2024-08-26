package main

import (
	"fmt"
	"net/url"
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

func filterContent(htmlContent, filter string) bool {
	return strings.Contains(htmlContent, filter)
}

func (cfg *config) crawlPage(rawCurrentURL, filter string) {

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

		if !filterContent(htmlBody, filter) {
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
			cfg.crawlPage(u, filter)
		}
	}()
}
