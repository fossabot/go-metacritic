package metacritic

import (
	"net/http"
	"sync"
)

// Client is the interface used by the Crawler to retrieve the data from the url.
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

// Result is the result returned from the DefaultCrawler.
type Result struct {
	Error    error
	Response *http.Response
}

// DefaultCrawler is the default implementation for the Crawler interface.
type DefaultCrawler struct {
	Concurrent int
	Client     Client
	UserAgent  string
}

func (c *DefaultCrawler) doQuery(url string) *Result {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return &Result{
			Error: err,
		}
	}
	req.Header.Set("User-Agent", c.UserAgent)

	res, err := c.Client.Do(req)
	if err != nil {
		return &Result{
			Error: err,
		}
	}

	return &Result{
		Response: res,
	}
}

// Crawl will start the crawling process for given urls in concurrent.
func (c *DefaultCrawler) Crawl(urls []string) []*Result {
	var mu sync.Mutex

	results := make([]*Result, 0, len(urls))

	sem := make(chan struct{}, c.Concurrent)
	for _, url := range urls {
		sem <- struct{}{}
		go func(u string) {
			result := c.doQuery(u)

			mu.Lock()
			defer mu.Unlock()

			results = append(results, result)

			<-sem
		}(url)
	}

	for i := 0; i < c.Concurrent; i++ {
		sem <- struct{}{}
	}

	return results
}

// CrawlOne calls Crawl returning the first element.
//
// Caution: Crawl is concurrent and is using a slice, so calling twice CrawlOne
// could produce different results.
func (c *DefaultCrawler) CrawlOne(url string) *Result {
	return c.Crawl([]string{url})[0]
}
