package metacritic

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/hbakhtiyor/strsim"
)

const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
	"(KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36"

// Game represents the result from metacritic.
type Game struct {
	Link      string
	MetaScore uint8
	Title     string
	UserScore float64
}

// Crawler is the interface used by the Metacritic struct to retrieve the data.
type Crawler interface {
	Crawl(urls []string) []*Result
	CrawlOne(url string) *Result
}

// Metacritic is the main service to get the details for a game.
type Metacritic struct {
	crawler Crawler
}

// New returns a new Metacritic given a Client, concurrent and useragent.
//
// It will use the the DefaultCrawler with given settings.
func New(c Client, concurrent int, ua string) *Metacritic {
	return NewWithCrawler(&DefaultCrawler{
		Client:     c,
		Concurrent: concurrent,
		UserAgent:  ua,
		Parser:     goquery.NewDocumentFromReader,
	})
}

// NewWithCrawler returns a new Metacritic given a custom Crawler implementation.
func NewWithCrawler(c Crawler) *Metacritic {
	return &Metacritic{
		crawler: c,
	}
}

// NewWithDefaults returns a new Metacritic with default settings.
//
// In most cases this is what you want.
func NewWithDefaults() *Metacritic {
	return New(
		&http.Client{Timeout: time.Second * 10},
		3,
		defaultUserAgent,
	)
}

// parseGamePage tries to find the scores for the given Result.
func (m *Metacritic) parseGamePage(r *Result) *Game {
	if r.Error != nil {
		return nil
	}

	title := r.Doc.Find("div.product_title a h1").Text()

	metaText := strings.TrimSpace(r.Doc.Find("div.metascore_w span").Text())
	if metaText == "" {
		metaText = "0"
	}
	metaNumber, err := strconv.Atoi(metaText)
	if err != nil {
		metaNumber = 0
	}
	meta := uint8(metaNumber)

	userscoreText := strings.TrimSpace(r.Doc.Find("div.userscore_wrap * div.metascore_w.user").Text())
	if userscoreText == "" || userscoreText == "tbd" {
		userscoreText = "0"
	}
	userscore, err := strconv.ParseFloat(userscoreText, 10)
	if err != nil {
		userscore = 0.0
	}

	return &Game{
		Title:     title,
		Link:      r.URL,
		MetaScore: meta,
		UserScore: userscore,
	}
}

// parseSearchPage tries to find the urls for the found games.
func (m *Metacritic) parseSearchPage(r *Result) []string {
	var urls []string

	r.Doc.Find("li.result").Each(func(i int, s *goquery.Selection) {
		if link, ok := s.Find("h3.product_title a").Attr("href"); ok {
			urls = append(urls, fmt.Sprintf(`https://www.metacritic.com%s`, link))
		}
	})

	return urls
}

// startSearch will start the crawling process of metacritic.
//
// It will call the search page with title and platform crawling for all the detail pages.
// Then it will crawl every detail page in concurrent to extract the scores.
func (m *Metacritic) startSearch(title string, platform Platform) ([]*Game, error) {
	var retVal []*Game

	result := m.crawler.CrawlOne(fmt.Sprintf(
		`https://www.metacritic.com/search/game/%s/results?plats[%s]=1&search_type=advanced`,
		url.PathEscape(title),
		platform,
	))
	if result == nil || result.Error != nil || result.Doc == nil {
		return retVal, fmt.Errorf("cannot crawl search result page")
	}

	urls := m.parseSearchPage(result)

	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, g := range m.crawler.Crawl(urls) {
		wg.Add(1)
		go func(g *Result) {
			defer wg.Done()
			game := m.parseGamePage(g)
			if game == nil {
				return
			}

			mu.Lock()
			defer mu.Unlock()
			retVal = append(retVal, game)
		}(g)
	}

	wg.Wait()

	return retVal, nil
}

// findBestMatch returns the best match for title for the given games.
func (m *Metacritic) findBestMatch(title string, games []*Game) *Game {
	l := len(games)
	if l == 0 {
		return nil
	}

	if l == 1 {
		return games[0]
	}

	tmp := make(map[string]*Game)
	titles := make([]string, 0, len(games))
	for _, g := range games {
		tmp[g.Title] = g
		titles = append(titles, g.Title)
	}

	match, _ := strsim.FindBestMatch(title, titles)
	if match == nil {
		return nil
	}

	return tmp[match.BestMatch.Target]
}

// Search will start the crawl and parse process for the given title and platform.
func (m *Metacritic) Search(title string, platform Platform) ([]*Game, error) {
	return m.startSearch(title, platform)
}

// SearchBestMatch will call Search and returns then the best match.
//
// The best match is calculated with the "Dice's Coefficient" by the
// using the external lib "https://github.com/hbakhtiyor/strsim".
func (m *Metacritic) SearchBestMatch(title string, platform Platform) *Game {
	games, err := m.Search(title, platform)
	if err != nil {
		return nil
	}

	return m.findBestMatch(title, games)
}
