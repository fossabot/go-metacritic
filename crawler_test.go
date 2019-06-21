package metacritic_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PuerkitoBio/goquery"

	"github.com/stahlstift/go-metacritic"
)

type MockClient struct {
	metacritic.Client

	DoFn       func(req *http.Request) (*http.Response, error)
	DoFnCalled bool
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	m.DoFnCalled = true

	if m.DoFn != nil {
		return m.DoFn(req)
	}

	return httptest.NewRecorder().Result(), nil
}

const userAgent = "unitTest"

func TestUserAgentIsAdded(t *testing.T) {
	t.Parallel()

	mock := &MockClient{}
	mock.DoFn = func(req *http.Request) (response *http.Response, err error) {
		if req.Header.Get("User-Agent") != userAgent {
			t.Fatalf("User-Agent was not added")
		}
		return httptest.NewRecorder().Result(), nil
	}
	c := &metacritic.DefaultCrawler{
		Client:     mock,
		Concurrent: 2,
		Parser:     goquery.NewDocumentFromReader,
		UserAgent:  userAgent,
	}
	_ = c.Crawl([]string{"http://www.example.org", "http://www.example.org/1"})
}

func TestBrokenURL(t *testing.T) {
	t.Parallel()

	mock := &MockClient{}
	c := &metacritic.DefaultCrawler{
		Client:     mock,
		Concurrent: 2,
		Parser:     goquery.NewDocumentFromReader,
		UserAgent:  userAgent,
	}

	res := c.Crawl([]string{"invalid_url:678888"})
	if len(res) == 0 {
		t.Fatal("Crawler() returned 0 results")
	}

	if res[0].Error == nil {
		t.Fatal("Expected error - received nil")
	}
}

func TestParserError(t *testing.T) {
	t.Parallel()

	parseError := fmt.Errorf("parse_error")

	mock := &MockClient{}
	c := &metacritic.DefaultCrawler{
		Client:     mock,
		Concurrent: 2,
		Parser: func(r io.Reader) (document *goquery.Document, err error) {
			return nil, parseError
		},
		UserAgent: userAgent,
	}

	res := c.Crawl([]string{"http://www.example.org", "http://www.example.org/1"})
	if len(res) == 0 {
		t.Fatal("Crawler() returned 0 results")
	}

	if res[0].Error != parseError {
		t.Fatalf("Crawler() did not receive correct error. Expected '%s' - got '%s'", parseError, res[0].Error)
	}
}

func TestCrawlClientError(t *testing.T) {
	t.Parallel()

	mockError := fmt.Errorf("unittest_error")

	mock := &MockClient{}
	mock.DoFn = func(req *http.Request) (response *http.Response, err error) {
		return nil, mockError
	}
	c := &metacritic.DefaultCrawler{
		Client:     mock,
		Concurrent: 2,
		Parser:     goquery.NewDocumentFromReader,
		UserAgent:  userAgent,
	}
	res := c.Crawl([]string{"http://www.example.org", "http://www.example.org/1"})
	if len(res) == 0 {
		t.Fatal("Crawler() returned 0 results")
	}

	for _, r := range res {
		if r.Error != mockError {
			t.Fatalf("Crawler() did not receive correct error. Expected '%s' - got '%s'", mockError, r.Error)
		}
	}

}

func TestCrawlOne(t *testing.T) {
	t.Parallel()

	mock := &MockClient{}
	c := &metacritic.DefaultCrawler{
		Client:     mock,
		Concurrent: 2,
		Parser:     goquery.NewDocumentFromReader,
		UserAgent:  userAgent,
	}
	res := c.CrawlOne("http://www.example.org")
	if res == nil {
		t.Fatal("CrawlOne() did no returned a result")
	}
}
