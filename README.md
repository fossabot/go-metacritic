# go-metacritic
[![build status](https://secure.travis-ci.org/stahlstift/go-metacritic.svg?branch=master)](http://travis-ci.org/stahlstift/go-metacritic) [![GoDoc](https://godoc.org/github.com/stahlstift/go-metacritic?status.png)](http://godoc.org/github.com/stahlstift/go-metacritic) [![Sourcegraph Badge](https://sourcegraph.com/github.com/stahlstift/go-metacritic/-/badge.svg)](https://sourcegraph.com/github.com/stahlstift/go-metacritic?badge)

go-metacritic is a simple lib to crawl Metacritic for the Metascore and userscore of a game and platform.

## Changelog

*    **v0.1.0** : Initial release.

## Example

```Go
package main

import (
    "fmt"
    
    "github.com/stahlstift/go-metacritic/metacritic"
)

func main() {
    mc := metacritic.NewWithDefaults()
    res, err := mc.Search("Mario", metacritic.Switch)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Search results for %s on Platform %s\n\n", "Mario", metacritic.Switch)
    for _, game := range res {
        fmt.Printf("%+v\n", game)
    }

    game := mc.SearchBestMatch("Mario Kart 8", metacritic.Switch)
    if game != nil {
        fmt.Printf("BestMatch result for %s on Platform %s\n\n", "Mario Kart 8", metacritic.Switch)
        fmt.Printf("%+v\n", game)
    }
}
```
