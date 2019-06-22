package metacritic

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/hbakhtiyor/strsim"
)

// Game represents the result from metacritic.
type Game struct {
	Link      string
	MetaScore uint8
	Title     string
	UserScore float32
}

// Crawler is the interface used by the Metacritic struct to retrieve the data.
type Crawler interface {
	Crawl(urls []string) []*Result
	CrawlOne(url string) *Result
}

type Parser interface {
	Game(body io.Reader) *Game
	Search(body io.Reader) []string
}

// Metacritic is the main service to get the details for a game.
type Metacritic struct {
	Crawler Crawler
	Parser  Parser
}

// New returns a new Metacritic given a Client, concurrent and useragent.
//
// It will use the the DefaultCrawler with given settings.
func New() *Metacritic {
	return &Metacritic{
		Crawler: &DefaultCrawler{
			Client:     &http.Client{Timeout: time.Second * 10},
			Concurrent: 3,
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
				"(KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36",
		},
		Parser: &DefaultParser{},
	}
}

// startSearch will start the crawling process of metacritic.
//
// It will call the search page with title and platform crawling for all the detail pages.
// Then it will crawl every detail page in concurrent to extract the scores.
func (m *Metacritic) startSearch(title string, platform Platform) ([]*Game, error) {
	var retVal []*Game

	result := m.Crawler.CrawlOne(fmt.Sprintf(
		`https://www.metacritic.com/search/game/%s/results?plats[%s]=1&search_type=advanced`,
		url.PathEscape(title),
		platform,
	))
	if result == nil || result.Error != nil {
		return retVal, fmt.Errorf("cannot crawl search result page")
	}

	defer result.Response.Body.Close()
	urls := m.Parser.Search(result.Response.Body)

	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, g := range m.Crawler.Crawl(urls) {
		wg.Add(1)
		go func(g *Result) {
			defer wg.Done()

			if g == nil || g.Error != nil {
				return
			}

			defer g.Response.Body.Close()
			game := m.Parser.Game(g.Response.Body)
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
