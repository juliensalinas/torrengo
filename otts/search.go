// Package otts searches and downloads torrents from 1337x.to
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
// - DescURL: the torrent description page
//
// - Name: the torrent name
//
// - Size: the size of the file to be downloaded
//
// - UplDate: the date of upload
//
// - Leechers: the number of leechers (set to -1 if cannot be converted to integer)
//
// - Seechers: the number of seechers (set to -1 if cannot be converted to integer)
//
// Magnet file extraction are achieved by ExtractMag().
// Input is the url of the torrent page.
// Output is the magnet link.

package otts

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/juliensalinas/torrengo/core"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

const baseURL string = "https://1337x.to"

// Torrent contains meta information about the torrent
type Torrent struct {
	DescURL string
	Name    string
	Size    string
	UplDate string
	// Seeders and Leechers are converted to -1 if cannot be converted to integers
	Seeders  int
	Leechers int
}

// A typical final url looks like:
// https://1337x.to/search/Dumas/1/
func buildSearchURL(in string) (string, error) {
	var URL *url.URL
	URL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("error during url parsing: %v", err)
	}

	URL.Path += "/search/" + in + "/1/"

	return URL.String(), nil
}

func parseSearchPage(r io.Reader) ([]Torrent, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("could not load html response into GoQuery: %v", err)
	}

	// torrents stores a list of torrents made up of the torrent description url,
	// its name, its size, its upload date, its seeders, and its leechers
	var torrents []Torrent

	// Results are located in a clean html <table>
	doc.Find("tbody tr").Each(func(i int, s *goquery.Selection) {
		var t Torrent

		// Name is the text of the 2nd <a> tag, and desc URL is the href
		path, ok := s.Find("a").Eq(1).First().Attr("href")
		if !ok {
			log.Debug("Could not find a description page for a torrent so ignoring it")
			return
		}
		t.DescURL = baseURL + path
		t.Name = s.Find("a").Eq(1).First().Text()

		// Seeders and leechers are located in the 2nd and 3rd <td>.
		// We convert it to integers and if conversion fails we convert it to -1.
		seedersStr := s.Find("td").Eq(1).First().Text()
		seeders, err := strconv.Atoi(seedersStr)
		if err != nil {
			seeders = -1
		}
		t.Seeders = seeders

		leechersStr := s.Find("td").Eq(2).First().Text()
		leechers, err := strconv.Atoi(leechersStr)
		if err != nil {
			leechers = -1
		}
		t.Leechers = leechers

		// Upload date is the text of the 4th <td> tag
		t.UplDate = s.Find("td").Eq(3).First().Text()

		// Size is the text of the 5th <td> tag
		t.Size = s.Find("td").Eq(4).First().Text()

		torrents = append(torrents, t)
	})

	return torrents, nil
}

// Lookup takes a user search as a parameter, launches the http request
// with a custom timeout, and returns clean torrent information fetched from 1337x.to
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
