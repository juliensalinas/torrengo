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
// - Magnet: the torrent magnet
// - Name: the torrent name
// - Size: the size of the file to be downloaded
// - UplDate: the date of upload
// - Leechers: the number of leechers (set to -1 if cannot be converted to integer)
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

	"github.com/PuerkitoBio/goquery"
)

const baseURL string = "https://1337x.to"
const userAgent string = "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.62 Safari/537.36"

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

func fetch(url string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %v", err)
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not launch request: %v", err)
	}

	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	return resp, nil
}

func parseSearchPage(r io.Reader) ([]Torrent, error) {
	// Load html response into GoQuery
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("could not load html response into GoQuery: %v", err)
	}

	// torrents stores a list of torrents made up of the torrent description url,
	// its name, its size, its seeders, and its leechers
	var torrents []Torrent

	// Results are located in a clean html <table>
	doc.Find("tbody tr").Each(func(i int, s *goquery.Selection) {
		var t Torrent

		// Magnet is the href of the 4th <a> tag
		s.Find("a").Eq(1).Each(func(i int, ss *goquery.Selection) {
			t.Name = ss.Text()
			descURL, ok := ss.Attr("href")
			if ok {
				t.DescURL = descURL
			}
		})

		// Seeders and leechers are located in the 3rd and 4th <td>.
		// We convert it to integers and if conversion fails we convert it to -1.
		s.Find("td").Eq(1).Each(func(i int, ss *goquery.Selection) {
			seedersStr := ss.Text()
			seeders, err := strconv.Atoi(seedersStr)
			if err != nil {
				seeders = -1
			}
			t.Seeders = seeders
		})
		s.Find("td").Eq(2).Each(func(i int, ss *goquery.Selection) {
			leechersStr := ss.Text()
			leechers, err := strconv.Atoi(leechersStr)
			if err != nil {
				leechers = -1
			}
			t.Leechers = leechers
		})

		s.Find("td").Eq(3).Each(func(i int, ss *goquery.Selection) {
			t.UplDate = ss.Text()
		})

		s.Find("td").Eq(4).Each(func(i int, ss *goquery.Selection) {
			t.Size = ss.Text()
		})

		torrents = append(torrents, t)
	})

	return torrents, nil
}

// Lookup takes a user search as a parameter and
// returns clean torrent information fetched from 1337x.to
func Lookup(in string) ([]Torrent, error) {

	url, err := buildSearchURL(in)
	if err != nil {
		return nil, fmt.Errorf("error while building url: %v", err)
	}

	resp, err := fetch(url)
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
