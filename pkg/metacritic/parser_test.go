package metacritic

import (
	"os"
	"testing"
)

func TestParseSearchPage(t *testing.T) {
	t.Parallel()

	file, err := os.Open("./testdata/search_result.html")
	if err != nil {
		t.Fatalf("error opening './testdata/search_result.html' ('%s')", err)
	}

	p := &DefaultParser{}
	urls := p.Search(file)
	if len(urls) != 2 {
		t.Fatalf("error parsing game urls")
	}
}

func TestParseGamePage(t *testing.T) {
	t.Parallel()

	file, err := os.Open("./testdata/mario_party.html")
	if err != nil {
		t.Fatalf("error opening './testdata/mario_party.html' ('%s')", err)
	}

	p := &DefaultParser{}
	game := p.Game(file)
	if game.Title != "Super Mario Party" {
		t.Fatalf("wrong game '%s' returned", game.Title)
	}

	if game.MetaScore != 76 {
		t.Fatalf("wrong metascore '%d' returned", game.MetaScore)
	}

	if game.UserScore != 7.5 {
		t.Fatalf("wrong userscore '%f' returned", game.UserScore)
	}
}
