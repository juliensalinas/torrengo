// Package ygg searches and downloads torrents from Ygg Torrent
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
package ygg

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

const baseURL = "yggtorrent.to"

// searchURL is the url used to retrieve a list of torrents based on user keywords.
// A typical final url looks like:
// https://www.yggtorrent.is/engine/search?name=alexandre+dumas&do=search
var searchURL = url.URL{
	Scheme: "https",
	Host:   baseURL,
	Path:   "engine/search",
}

// searchParams are the hardcoded GET parameters.
// Some other dynamic params are added further in the program.
var searchParams = url.Values{
	"do": {"search"},
}

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

func parseSearchPage(r io.Reader) ([]Torrent, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("could not load html response into GoQuery: %v", err)
	}

	// torrents stores a list of torrents made up of the torrent description url,
	// its name, its size, its upload date, its seeders, and its leechers
	var torrents []Torrent

	// Results are located in a clean html <table> whose class is table
	doc.Find(".table tbody tr").Each(func(i int, s *goquery.Selection) {
		var t Torrent

		// Torrent name is the text of the 2th <td> tag and descURL is its href
		descURL, ok := s.Find("td a").Eq(1).First().Attr("href")
		if !ok {
			log.Debug("Could not find description URL for a torrent so ignoring it")
			return
		}
		t.DescURL = descURL

		t.Name = s.Find("td a").Eq(1).First().Text()

		// Upload date is the text of the div whose class is hidden in the 3rd <td> tag.
		// A proper timestamp is retrieved. We convert it to datetime.
		timestampStr := s.Find("td").Eq(4).First().Find(".hidden").First().Text()
		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			t.UplDate = ""
		} else {
			t.UplDate = time.Unix(timestamp, 0).Format("2006/01/02 15:04")
		}

		// File size is the text of the 4th <td> tag
		t.Size = s.Find("td").Eq(5).First().Text()

		// Seeders is the text of the 6th <td> tag
		seedersStr := s.Find("td").Eq(7).First().Text()
		seeders, err := strconv.Atoi(seedersStr)
		if err != nil {
			seeders = -1
		}
		t.Seeders = seeders

		// Leechers is the text of the 7th <td> tag
		leechersStr := s.Find("td").Eq(8).First().Text()
		leechers, err := strconv.Atoi(leechersStr)
		if err != nil {
			leechers = -1
		}
		t.Leechers = leechers

		torrents = append(torrents, t)
	})

	return torrents, nil
}

// Lookup takes a user search as a parameter, launches the http request
// with a custom timeout, and returns clean torrent information fetched from Ygg Torrent
func Lookup(in string, timeout time.Duration) ([]Torrent, error) {
	client := &http.Client{
		Timeout: timeout,
	}

	searchParams.Add("name", in)
	searchURL.RawQuery = searchParams.Encode()

	resp, err := core.Fetch(searchURL.String(), client)
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
