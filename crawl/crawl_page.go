package crawl

import (
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/itxtx/crawler_go/config"
	"github.com/itxtx/crawler_go/extractor"
)

type Config struct {
	Links              []extractor.LinkInfo
	Pages              map[string]int
	BaseURL            *url.URL
	Mu                 *sync.Mutex
	ConcurrencyControl chan struct{}
	Wg                 *sync.WaitGroup
	MaxPages           int
	CrawlerConfig      *config.CrawlerConfig
}

type LinkInfo struct {
	URL         string
	Description string
}

func (cfg *Config) AddLink(link extractor.LinkInfo) bool {
	cfg.Mu.Lock()
	defer cfg.Mu.Unlock()

	if len(cfg.Links) >= cfg.MaxPages {
		return false
	}

	cfg.Links = append(cfg.Links, link)
	return true
}

type job struct {
	url    string
	filter string
}

func (cfg *Config) CrawlWorker(jobs <-chan job, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := range jobs {
		cfg.ProcessURL(j.url, j.filter)
	}
}

func (cfg *Config) CrawlWithWorkerPool(initialURL, filter string) {
	jobs := make(chan job, cfg.CrawlerConfig.MaxConcurrency)
	var wg sync.WaitGroup

	// Start worker pool
	for i := 0; i < cfg.CrawlerConfig.MaxConcurrency; i++ {
		wg.Add(1)
		go cfg.CrawlWorker(jobs, &wg)
	}

	// Add initial job
	jobs <- job{url: initialURL, filter: filter}

	// Close jobs channel when all work is done
	go func() {
		wg.Wait()
		close(jobs)
	}()

	// Wait for all jobs to complete
	wg.Wait()
}

func (cfg *Config) ProcessURL(rawCurrentURL, filter string) {
	cfg.Mu.Lock()
	if len(cfg.Pages) >= cfg.MaxPages {
		cfg.Mu.Unlock()
		return
	}
	cfg.Mu.Unlock()

	currentURL, err := url.Parse(rawCurrentURL)
	if err != nil {
		fmt.Println("Error parsing current URL:", err)
		return
	}

	htmlBody, err := fetchContent(currentURL.String())
	if err != nil {
		fmt.Println("Error fetching URL:", err)
		return
	}

	links, err := extractor.ExtractLinksAndDescriptions(htmlBody, currentURL, filter)
	if err != nil {
		fmt.Println("Error extracting links:", err)
		return
	}

	for _, link := range links {
		if cfg.AddLink(link) {
			fmt.Printf("Found matching link: %s\n", link.URL)
		} else {
			return
		}
	}

	if len(cfg.CrawlerConfig.Selectors) > 0 {
		extractor.ExtractMultipleContents(htmlBody, cfg.CrawlerConfig.Selectors, cfg.CrawlerConfig.SelectorType, cfg.CrawlerConfig.OutputFormat, true)
		extractor.ExtractContent(htmlBody, strings.Join(cfg.CrawlerConfig.Selectors, ","), cfg.CrawlerConfig.SelectorType, cfg.CrawlerConfig.OutputFormat, true)
	}

	urls, err := getURLsFromHTML(htmlBody, currentURL.String())
	if err != nil {
		fmt.Println("Error extracting URLs:", err)
		return
	}

	for _, u := range urls {
		cfg.CrawlWithWorkerPool(u, filter)
	}
}
