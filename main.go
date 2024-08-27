package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/itxtx/crawler_go/config"
	"github.com/itxtx/crawler_go/crawl"
	"github.com/itxtx/crawler_go/extractor"
)

func main() {
	configArgs, err := config.ParseArgs(os.Args)
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

	cfg := &crawl.Config{
		Links:              make([]extractor.LinkInfo, 0, configArgs.MaxPages),
		Pages:              make(map[string]int),
		BaseURL:            configArgs.BaseURL,
		Mu:                 &sync.Mutex{},
		ConcurrencyControl: make(chan struct{}, configArgs.MaxConcurrency),
		Wg:                 &sync.WaitGroup{},
		MaxPages:           configArgs.MaxPages,
		CrawlerConfig:      configArgs,
	}

	// Start crawling with the new worker pool approach
	cfg.CrawlWithWorkerPool(configArgs.BaseURL.String(), configArgs.Filter)

	fmt.Println("\nCrawl results:")
	for _, link := range cfg.Links {
		fmt.Printf("URL: %s\nDescription: %s\n\n", link.URL, link.Description)
	}

	config.PrintReport(cfg.Pages, configArgs.BaseURL.String())
}
