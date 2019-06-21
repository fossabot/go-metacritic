package metacritic

import (
	"io"
	"net/http"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

// Client is the interface used by the Crawler to retrieve the data from the url.
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

// Result is the result returned from the DefaultCrawler.
type Result struct {
	Doc   *goquery.Document
	URL   string
	Error error
}

// DefaultCrawler is the default implementation for the Crawler interface.
type DefaultCrawler struct {
	Concurrent int
	Client     Client
	Parser     func(r io.Reader) (*goquery.Document, error)
	UserAgent  string
}

func (c *DefaultCrawler) doQuery(url string) (*goquery.Document, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.UserAgent)

	return c.doQueryRequest(req)
}

func (c *DefaultCrawler) doQueryRequest(req *http.Request) (*goquery.Document, error) {
	res, err := c.Client.Do(req)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	doc, err := c.Parser(res.Body)
	if err != nil {
		return nil, err
	}

	return doc, err
}

// Crawl will start the crawling process for given urls in concurrent.
func (c *DefaultCrawler) Crawl(urls []string) []*Result {
	var mu sync.Mutex

	results := make([]*Result, 0, len(urls))

	sem := make(chan struct{}, c.Concurrent)
	for _, url := range urls {
		sem <- struct{}{}
		go func(u string) {
			doc, err := c.doQuery(u)

			mu.Lock()
			defer mu.Unlock()

			results = append(results, &Result{
				Doc:   doc,
				Error: err,
				URL:   u,
			})

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
