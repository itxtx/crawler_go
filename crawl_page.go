package main

import (
	"fmt"
	"net/url"
	"strings"
)

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

		// Get the HTML from the current URL
		fmt.Printf("Crawling: %s\n", currentURL)
		htmlBody, err := fetchContent(currentURL.String())
		if err != nil {
			fmt.Println("Error fetching URL:", err)
			return
		}

		// Extract links and descriptions from the HTML
		links, err := extractLinksAndDescriptions(htmlBody, currentURL, filter)
		if err != nil {
			fmt.Println("Error extracting links:", err)
			return
		}

		// Add matching links to the list
		for _, link := range links {
			if cfg.addLink(link) {
				fmt.Printf("Found matching link: %s\n", link.URL)
			} else {
				return // Stop if we've reached the maximum number of links
			}
		}

		// Extract content based on selectors if provided
		if len(cfg.crawlerConfig.Selectors) > 0 {
			extractMultipleContents(htmlBody, cfg.crawlerConfig.Selectors, cfg.crawlerConfig.SelectorType, cfg.crawlerConfig.OutputFormat, true)
			extractContent(htmlBody, strings.Join(cfg.crawlerConfig.Selectors, ","), cfg.crawlerConfig.SelectorType, cfg.crawlerConfig.OutputFormat, true)
		}

		// Extract all URLs from the HTML for further crawling
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
