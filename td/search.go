// Package td searches and downloads torrents from torrentdownloads.me
//
// No check is done here regarding the user input. This check should be
// achieved by the caller.
// Parsing is achieved thanks to the GoQuery library.
// Comments common to all scraping libs are already done in the arc package which is very
// similar to this package. Only additional comments specific to this lib are present here.
//
// Torrent search is achieved by Lookup().
// Input is a search string.
// Output is a slice of maps made up of the following keys:
//
// - DescUrl: the torrent description dedicated url
//
// - Name: the torrent name
//
// - Size: the size of the file to be downloaded
//
// - Leechers: the number of leechers (set to -1 if cannot be converted to integer)
//
// - Seechers: the number of seechers (set to -1 if cannot be converted to integer)
//
// Torrent url and magnet file extraction are achieved by ExtractTorAndMag().
// Input is the url of the torrent page.
// Output are the torrent url and the magnet link.
//
// Download of torrent file is achieved by DlFile() (very tricky
// because torrentdownloads has a Cloudflare protection so  does not work 100% of the time).
// Input is the torrent file url.
// Output is the local path where the torrent file was downloaded.

package td

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/juliensalinas/torrengo/core"
	log "github.com/sirupsen/logrus"
)

const baseURL string = "https://www.torrentdownloads.me"

// Torrent contains meta information about the torrent
type Torrent struct {
	DescURL string
	Name    string
	Size    string
	// Seeders and Leechers are converted to -1 if cannot be converted to integers
	Seeders  int
	Leechers int
}

// A typical final url looks like:
// https://www.torrentdownloads.me/search/?search=Dumas
func buildSearchURL(in string) (string, error) {
	var URL *url.URL
	URL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("error during url parsing: %v", err)
	}

	URL.Path += "/search"

	params := url.Values{}
	params.Add("search", in)
	URL.RawQuery = params.Encode()

	return URL.String(), nil
}

func parseSearchPage(r io.Reader) ([]Torrent, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("could not load html response into GoQuery: %v", err)
	}

	// torrents stores a list of torrents made up of the torrent description url,
	// its name, its size, its seeders, and its leechers
	var torrents []Torrent

	// Get the total number of items found
	l := doc.Find(".inner_container ").Children().Size()

	doc.Find(".inner_container ").Children().Each(func(i int, s *goquery.Selection) {
		var t Torrent
		// Many elements in inner_container are junk (ads, empty stuffs,...) so
		// we only take elements between 10 and 2 before the end
		if i > 9 && i < l-2 {
			// Get path to torrent description page from a "<a>" tag located
			// inside a <p> tag
			path, ok := s.Find("p a").First().Attr("href")
			if !ok {
				log.Debug("Could not find the description URL of a torrent")
				return
			}
			t.DescURL = baseURL + path

			// Get name from the same place as path
			t.Name = strings.TrimSpace(s.Find("p a").First().Text())

			// Get leechers, seeders and size from <span> tags 2, 3 and 4.
			// Try to convert leechers and seeders to integers but if does not work
			// we do not stop for all that: we just set the leechers and seeders to
			// -1 so the calling library can differentiate it.
			leechersStr := s.Find("span").Eq(1).First().Text()
			leechers, err := strconv.Atoi(leechersStr)
			if err != nil {
				leechers = -1
			}
			t.Leechers = leechers

			seedersStr := s.Find("span").Eq(2).First().Text()
			seeders, err := strconv.Atoi(seedersStr)
			if err != nil {
				seeders = -1
			}
			t.Seeders = seeders

			size := s.Find("span").Eq(3).First().Text()
			t.Size = size

			torrents = append(torrents, t)
		}
	})

	return torrents, nil
}

// Lookup takes a user search as a parameter, launches the http request
// with a custom timeout, and returns clean torrent information fetched from torrentdownloads.me
func Lookup(in string, timeout time.Duration) ([]Torrent, error) {
	client := &http.Client{
		Timeout: timeout,
	}

	url, err := buildSearchURL(in)
	if err != nil {
		return nil, fmt.Errorf("error while building url: %v", err)
	}

	resp, err := core.Fetch(url, client)
	if err != nil {
		return nil, fmt.Errorf("error while fetching url: %v", err)
	}
	defer resp.Body.Close()

	torrents, err := parseSearchPage(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error while parsing torrent search results: %v", err)
	}

	return torrents, nil
}
