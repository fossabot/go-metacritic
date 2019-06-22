// +build integration

package metacritic_test

import (
	"testing"

	"github.com/stahlstift/go-metacritic/pkg/metacritic"
)

func TestMetacritic_HTMLisUnchanged(t *testing.T) {
	t.Parallel()

	mc := metacritic.New()
	game := mc.SearchBestMatch("Mario Kart 8", metacritic.Switch)
	if game == nil {
		t.Fatalf("No result - maybe metacritic changed the html")
	}

	if game.Title != "Mario Kart 8 Deluxe" {
		t.Fatalf("Returned wrong title '%s' - maybe metacritic changed the html", game.Title)
	}

	if game.UserScore < 1 {
		t.Fatalf("Returned wrong userscore '%f' - maybe metacritic changed the html", game.UserScore)
	}

	if game.MetaScore < 50 {
		t.Fatalf("Returned wrong metascore '%d' - maybe metacritic changed the html", game.MetaScore)
	}
}
