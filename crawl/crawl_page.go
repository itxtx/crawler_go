package crawl

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

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

func (cfg *Config) CrawlWorker(jobs <-chan job, results chan<- []job) {
	for j := range jobs {
		newJobs := cfg.ProcessURL(j.url, j.filter)
		results <- newJobs
	}
}

func (cfg *Config) CrawlWithWorkerPool(initialURL, filter string) {
	jobs := make(chan job, cfg.CrawlerConfig.MaxConcurrency)
	results := make(chan []job, cfg.CrawlerConfig.MaxConcurrency)

	// Start worker pool
	for i := 0; i < cfg.CrawlerConfig.MaxConcurrency; i++ {
		go cfg.CrawlWorker(jobs, results)
	}

	// Add initial job
	jobs <- job{url: initialURL, filter: filter}

	activeJobs := 1
	processedURLs := make(map[string]bool)

	for activeJobs > 0 {
		select {
		case newJobs := <-results:
			activeJobs--
			for _, j := range newJobs {
				if !processedURLs[j.url] {
					jobs <- j
					activeJobs++
					processedURLs[j.url] = true
				}
			}
		}

		if activeJobs == 0 {
			close(jobs)
		}
	}

	close(results)
	fmt.Println("All jobs processed.")
}

func (cfg *Config) ProcessURL(rawCurrentURL, filter string) []job {
	cfg.Mu.Lock()
	if len(cfg.Pages) >= cfg.MaxPages {
		cfg.Mu.Unlock()
		return []job{} // Return an empty slice instead of nil
	}
	cfg.Mu.Unlock()

	currentURL, err := url.Parse(rawCurrentURL)
	if err != nil {
		fmt.Println("Error parsing current URL:", err)
		return []job{} // Return an empty slice
	}

	htmlBody, err := fetchContent(currentURL.String())
	if err != nil {
		fmt.Println("Error fetching URL:", err)
		return []job{} // Return an empty slice
	}

	links, err := extractor.ExtractLinksAndDescriptions(htmlBody, currentURL, filter)
	if err != nil {
		fmt.Println("Error extracting links:", err)
		return []job{} // Return an empty slice
	}

	for _, link := range links {
		if cfg.AddLink(link) {
			fmt.Printf("Found matching link: %s\n", link.URL)
		} else {
			return []job{} // Return an empty slice if we've reached the max pages
		}
	}

	if len(cfg.CrawlerConfig.Selectors) > 0 {
		extractor.ExtractMultipleContents(htmlBody, cfg.CrawlerConfig.Selectors, cfg.CrawlerConfig.SelectorType, cfg.CrawlerConfig.OutputFormat, true)
		extractor.ExtractContent(htmlBody, strings.Join(cfg.CrawlerConfig.Selectors, ","), cfg.CrawlerConfig.SelectorType, cfg.CrawlerConfig.OutputFormat, true)
	}

	if cfg.CrawlerConfig.ExtractVideos {
		// The printContent flag should be true since we want to print when ExtractVideos is enabled
		// The printing is handled inside ExtractVideos function now, similar to ExtractContent
		var vids []extractor.VideoInfo
		var err error

		// Use JavaScript engine if enabled
		if cfg.CrawlerConfig.EnableJS {
			timeout := time.Duration(cfg.CrawlerConfig.JSTimeout) * time.Second
			vids, err = extractor.ExtractVideosWithJavaScript(rawCurrentURL, currentURL, cfg.CrawlerConfig, timeout)
			if err != nil {
				fmt.Printf("JavaScript video extraction failed for %s, falling back to regular extraction: %v\n", rawCurrentURL, err)
				// Fallback to regular extraction
				vids, _ = extractor.ExtractVideos(htmlBody, currentURL, cfg.CrawlerConfig, true)
			} else if len(vids) > 0 {
				fmt.Printf("\nVideos found on %s using JavaScript extraction:\n", rawCurrentURL)
				for i, video := range vids {
					fmt.Printf("Video %d: %s (Platform: %s, Title: %s)\n", i+1, video.URL, video.Platform, video.Title)
				}
			}
		} else {
			// Regular video extraction
			vids, _ = extractor.ExtractVideos(htmlBody, currentURL, cfg.CrawlerConfig, true)
		}
		_ = vids // Store videos for future use if needed
	}

	urls, err := getURLsFromHTML(htmlBody, currentURL.String())
	if err != nil {
		fmt.Println("Error extracting URLs:", err)
		return []job{} // Return an empty slice
	}

	var newJobs []job
	for _, u := range urls {
		newJobs = append(newJobs, job{url: u, filter: filter})
	}

	if len(newJobs) > cfg.MaxPages-len(cfg.Pages) {
		newJobs = newJobs[:cfg.MaxPages-len(cfg.Pages)]
	}

	return newJobs
}
