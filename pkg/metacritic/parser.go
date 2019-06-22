package metacritic

import (
	"encoding/json"
	"io"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

type parsedGame struct {
	Context         string `json:"@context"`
	Type            string `json:"@type"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	URL             string `json:"url"`
	AggregateRating struct {
		Type        string `json:"@type"`
		BestRating  string `json:"bestRating"`
		WorstRating string `json:"worstRating"`
		RatingValue string `json:"ratingValue"`
		RatingCount string `json:"ratingCount"`
	} `json:"aggregateRating"`
	ContentRating string `json:"contentRating"`
}

type DefaultParser struct{}

// Search tries to find urls on the search result page.
func (p DefaultParser) Search(body io.Reader) []string {
	var urls []string

	tokenizer := html.NewTokenizer(body)

	found := false
	for {
		token := tokenizer.Next()

		if token == html.ErrorToken {
			break
		}

		if token == html.StartTagToken {
			if !found {
				h3 := tokenizer.Token()

				isH3 := h3.Data == "h3"
				if isH3 {
					for _, attr := range h3.Attr {
						if attr.Key == "class" && strings.Contains(attr.Val, "product_title") {
							found = true
						}
					}
				}
			}

			if found {
				a := tokenizer.Token()

				isA := a.Data == "a"
				if isA {
					for _, attr := range a.Attr {
						if attr.Key == "href" && strings.HasPrefix(attr.Val, "/game/") {
							urls = append(urls, "https://www.metacritic.com"+attr.Val)
							found = false
						}
					}

				}
			}

		}
	}

	return urls
}

func parseUserscore(tokenizer *html.Tokenizer) float32 {
	var userscore float32

	found := false
	for {
		if found {
			break
		}

		token := tokenizer.Next()

		if token == html.ErrorToken {
			break
		}

		if token == html.StartTagToken {
			div := tokenizer.Token()

			if div.Data == "div" {
				for _, attr := range div.Attr {
					if attr.Key == "class" &&
						strings.Contains(attr.Val, "metascore_w") &&
						strings.Contains(attr.Val, "user") &&
						strings.Contains(attr.Val, "game") {
						found = true

						tokenizer.Next()
						value := tokenizer.Token().Data
						value = strings.TrimSpace(value)

						t, err := strconv.ParseFloat(value, 10)
						if err == nil {
							userscore = float32(t)
						}
						break
					}
				}
			}
		}
	}

	return userscore
}

func parseJson(tokenizer *html.Tokenizer) *parsedGame {
	var parsedGame parsedGame

	stop := false
	found := false
	for {
		if stop {
			break
		}

		token := tokenizer.Next()

		if token == html.ErrorToken {
			break
		}

		if found && token == html.TextToken {
			stop = true
			err := json.Unmarshal(tokenizer.Text(), &parsedGame)
			if err != nil {
				return nil
			}
		}

		if token == html.StartTagToken {
			script := tokenizer.Token()

			isScript := script.Data == "script"
			if isScript {
				for _, attr := range script.Attr {
					if attr.Key == "type" && attr.Val == "application/ld+json" {
						found = true
						break
					}
				}
			}
		}
	}

	return &parsedGame
}

// Game tries to find the scores on the game detail page.
func (p DefaultParser) Game(body io.Reader) *Game {
	tokenizer := html.NewTokenizer(body)

	parsedGame := parseJson(tokenizer)
	if parsedGame == nil {
		return nil
	}

	metascore, _ := strconv.Atoi(parsedGame.AggregateRating.RatingValue)
	userscore := parseUserscore(tokenizer)

	return &Game{
		Link:      parsedGame.URL,
		Title:     parsedGame.Name,
		MetaScore: uint8(metascore),
		UserScore: userscore,
	}
}
