package metacritic_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stahlstift/go-metacritic/pkg/metacritic"
)

func ExampleMetacritic_Search() {
	mc := metacritic.New()
	res, err := mc.Search("Mario", metacritic.Switch)
	if err != nil {
		panic(err)
	}

	for _, game := range res {
		fmt.Printf("%d, %f, %s\n", game.MetaScore, game.UserScore, game.Title)
	}
}

func ExampleMetacritic_SearchBestMatch() {
	mc := metacritic.New()
	game := mc.SearchBestMatch("Mario Kart 8", metacritic.Switch)
	if game != nil {
		fmt.Printf("%d, %f, %s", game.MetaScore, game.UserScore, game.Title)
	}
}

func buildWithClient(c metacritic.Client) *metacritic.Metacritic {
	return &metacritic.Metacritic{
		Crawler: &metacritic.DefaultCrawler{
			Client:     c,
			Concurrent: 2,
			UserAgent:  "unittest",
		},
		Parser: &metacritic.DefaultParser{},
	}
}

var mockClient = &MockClient{
	DoFn: func(req *http.Request) (response *http.Response, err error) {
		res := httptest.NewRecorder().Result()

		if req.URL.String() == "https://www.metacritic.com/search/game/Mario/results?plats[268409]=1&search_type=advanced" {
			file, err := os.Open("./testdata/search_result.html")
			if err != nil {
				return nil, err
			}
			res.Body = file
		}

		if req.URL.String() == "https://www.metacritic.com/game/switch/super-mario-party" {
			file, err := os.Open("./testdata/mario_party.html")
			if err != nil {
				return nil, err
			}
			res.Body = file
		}

		if req.URL.String() == "https://www.metacritic.com/game/switch/super-mario-odyssey" {
			file, err := os.Open("./testdata/mario_odysee.html")
			if err != nil {
				return nil, err
			}
			res.Body = file
		}

		return res, nil
	},
}

func TestMetacritic_Search(t *testing.T) {
	t.Parallel()

	mc := buildWithClient(mockClient)

	res, err := mc.Search("Mario", metacritic.Switch)
	if err != nil {
		t.Fatalf("Search() returned an error '%s'", err)
	}

	if len(res) != 2 {
		t.Error("Search() did not return a correct result")
	}
}

func TestMetacritic_SearchNoGameResult(t *testing.T) {
	t.Parallel()

	mockClient := &MockClient{}
	mockClient.DoFn = func(req *http.Request) (response *http.Response, err error) {
		if req.URL.String() == "https://www.metacritic.com/search/game/Mario/results?plats[268409]=1&search_type=advanced" {
			res := httptest.NewRecorder().Result()

			file, err := os.Open("./testdata/search_result.html")
			if err != nil {
				return nil, err
			}
			res.Body = file
			return res, nil
		}

		return nil, fmt.Errorf("unittest")
	}

	mc := buildWithClient(mockClient)

	res, err := mc.Search("Mario", metacritic.Switch)
	if err != nil {
		t.Fatalf("Search() returned an error '%s'", err)
	}

	if len(res) != 0 {
		t.Error("Search() returned a result")
	}
}

func TestMetacritic_SearchNoSearchResult(t *testing.T) {
	t.Parallel()

	mockClient := &MockClient{}
	mockClient.DoFn = func(req *http.Request) (response *http.Response, err error) {
		return httptest.NewRecorder().Result(), nil
	}

	mc := buildWithClient(mockClient)

	res, err := mc.Search("Mario", metacritic.Switch)
	if err != nil {
		t.Fatalf("Search() returned an error '%s'", err)
	}

	if len(res) != 0 {
		t.Error("Search() returned a result")
	}
}

func TestMetacritic_SearchCrawlOneReturnsError(t *testing.T) {
	t.Parallel()

	mockClient := &MockClient{}
	mockClient.DoFn = func(req *http.Request) (response *http.Response, err error) {
		return nil, fmt.Errorf("unittest")
	}

	mc := buildWithClient(mockClient)

	_, err := mc.Search("Mario", metacritic.Switch)
	if err == nil {
		t.Error("Search() did not returned an error")
	}
}

func TestMetacritic_SearchBestMatch(t *testing.T) {
	t.Parallel()

	mc := buildWithClient(mockClient)

	res := mc.SearchBestMatch("Mario", metacritic.Switch)
	if res == nil {
		t.Fatalf("SearchBestMatch() did not return a result")
	}

	if res.Title != "Super Mario Party" {
		t.Fatalf("SearchBestMatch() returned '%s' instead of '%s'", res.Title, "Super Mario Party")
	}
}

func TestMetacritic_SearchBestMatchNoResult(t *testing.T) {
	t.Parallel()

	mockClient := &MockClient{}
	mockClient.DoFn = func(req *http.Request) (response *http.Response, err error) {
		res := httptest.NewRecorder().Result()

		if req.URL.String() == "https://www.metacritic.com/search/game/Mario/results?plats[268409]=1&search_type=advanced" {
			file, err := os.Open("./testdata/search_result_no_result.html")
			if err != nil {
				return nil, err
			}
			res.Body = file
		}

		return res, nil
	}

	mc := buildWithClient(mockClient)

	res := mc.SearchBestMatch("Mario", metacritic.Switch)
	if res != nil {
		t.Fatalf("SearchBestMatch() did returned a result")
	}
}

func TestMetacritic_SearchBestMatchOneResult(t *testing.T) {
	t.Parallel()

	mockClient := &MockClient{}
	mockClient.DoFn = func(req *http.Request) (response *http.Response, err error) {
		res := httptest.NewRecorder().Result()

		if req.URL.String() == "https://www.metacritic.com/search/game/Mario/results?plats[268409]=1&search_type=advanced" {
			file, err := os.Open("./testdata/search_result_one_game.html")
			if err != nil {
				return nil, err
			}
			res.Body = file
		}

		if req.URL.String() == "https://www.metacritic.com/game/switch/super-mario-party" {
			file, err := os.Open("./testdata/mario_party.html")
			if err != nil {
				return nil, err
			}
			res.Body = file
		}

		return res, nil
	}

	mc := buildWithClient(mockClient)

	res := mc.SearchBestMatch("Mario", metacritic.Switch)
	if res == nil {
		t.Fatalf("SearchBestMatch() did not return a result")
	}

	if res.Title != "Super Mario Party" {
		t.Fatalf("SearchBestMatch() returned '%s' instead of '%s'", res.Title, "Super Mario Party")
	}
}

func TestMetacritic_SearchBestMatchNoMetascore(t *testing.T) {
	t.Parallel()

	mockClient := &MockClient{}
	mockClient.DoFn = func(req *http.Request) (response *http.Response, err error) {
		res := httptest.NewRecorder().Result()

		if req.URL.String() == "https://www.metacritic.com/search/game/Mario/results?plats[268409]=1&search_type=advanced" {
			file, err := os.Open("./testdata/search_result_one_game2.html")
			if err != nil {
				return nil, err
			}
			res.Body = file
		}

		if req.URL.String() == "https://www.metacritic.com/game/switch/super-mario-odyssey" {
			file, err := os.Open("./testdata/mario_odysee_no_meta.html")
			if err != nil {
				return nil, err
			}
			res.Body = file
		}

		return res, nil
	}

	mc := buildWithClient(mockClient)

	res := mc.SearchBestMatch("Mario", metacritic.Switch)
	if res == nil {
		t.Fatalf("SearchBestMatch() did not return a result")
	}

	if res.MetaScore != 0 {
		t.Fatalf("Wrong Metascore returned '%d' instead of '0'", res.MetaScore)
	}
}

func TestMetacritic_SearchBestMatchNoUserscore(t *testing.T) {
	t.Parallel()

	mockClient := &MockClient{}
	mockClient.DoFn = func(req *http.Request) (response *http.Response, err error) {
		res := httptest.NewRecorder().Result()

		if req.URL.String() == "https://www.metacritic.com/search/game/Mario/results?plats[268409]=1&search_type=advanced" {
			file, err := os.Open("./testdata/search_result_one_game2.html")
			if err != nil {
				return nil, err
			}
			res.Body = file
		}

		if req.URL.String() == "https://www.metacritic.com/game/switch/super-mario-odyssey" {
			file, err := os.Open("./testdata/mario_odysee_no_user.html")
			if err != nil {
				return nil, err
			}
			res.Body = file
		}

		return res, nil
	}

	mc := buildWithClient(mockClient)

	res := mc.SearchBestMatch("Mario", metacritic.Switch)
	if res == nil {
		t.Fatalf("SearchBestMatch() did not return a result")
	}

	if res.UserScore != 0 {
		t.Fatalf("Wrong UserScore returned '%f' instead of '0'", res.UserScore)
	}
}

func TestMetacritic_SearchBestMatchWrongUserscore(t *testing.T) {
	t.Parallel()

	mockClient := &MockClient{}
	mockClient.DoFn = func(req *http.Request) (response *http.Response, err error) {
		res := httptest.NewRecorder().Result()

		if req.URL.String() == "https://www.metacritic.com/search/game/Mario/results?plats[268409]=1&search_type=advanced" {
			file, err := os.Open("./testdata/search_result_one_game2.html")
			if err != nil {
				return nil, err
			}
			res.Body = file
		}

		if req.URL.String() == "https://www.metacritic.com/game/switch/super-mario-odyssey" {
			file, err := os.Open("./testdata/mario_odysee_wrong_user.html")
			if err != nil {
				return nil, err
			}
			res.Body = file
		}

		return res, nil
	}

	mc := buildWithClient(mockClient)

	res := mc.SearchBestMatch("Mario", metacritic.Switch)
	if res == nil {
		t.Fatalf("SearchBestMatch() did not return a result")
	}

	if res.UserScore != 0 {
		t.Fatalf("Wrong UserScore returned '%f' instead of '0'", res.UserScore)
	}
}
