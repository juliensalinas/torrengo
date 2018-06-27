// Package arc searches archive.org and returns a clean list of torrents found
// on the first page.
// No check is done here regarding the user input. This check should be
// achieved by the caller.
// Parsing is achieved thanks to the GoQuery library.
//
// Input passed to the Search() function is a search string.
// Output is a slice of maps made up of 2 keys:
// descUrl: the torrent description dedicated url
// name: the torrent name
package arc

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const baseURL string = "https://archive.org"
const userAgent string = "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.62 Safari/537.36"

// Torrent contains meta information about the torrent
type Torrent struct {
	// Description url containing more info about the torrent including the torrent file address
	DescURL string
	Name    string
}

// buildURL encodes the user search keywords into a proper url.
// A typical final url looks like:
// https://archive.org/search.php?query=Dumas%20AND%20format%3A%22Archive%20BitTorrent%22
func buildURL(in string) (string, error) {
	// Add the following suffix to the query in order for archive.org
	// to return torrents only
	in += ` AND format:"Archive BitTorrent"`

	// Encode baseURL as an url.URL type (Parse expects a pointer)
	// so we can work on it more easily
	var URL *url.URL
	URL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("error during url parsing: %v", err)
	}

	// Create base path of URL
	URL.Path += "/search.php"

	// Add GET parameters
	params := url.Values{}
	params.Add("query", in)
	URL.RawQuery = params.Encode()

	return URL.String(), nil
}

// fetch opens a url and returns the resulting html page.
// Cannot use the straight http.Get function because need to
// modify headers in order to set a fake user-agent.
func fetch(url string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %v", err)
	}

	// Set the fake user agent
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

// parse parses an html slice of bytes and returns a clean list
// of torrents found in this page
func parse(r io.Reader) ([]Torrent, error) {

	// Load html response into GoQuery
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("could not load html response into GoQuery: %v", err)
	}

	// torrents stores a list of torrents made up of the torrent description url
	// and its name
	var torrents []Torrent

	doc.Find(".results ").Children().Each(func(i int, s *goquery.Selection) {
		// Get path to torrent description page from a "<a>" tag located inside a
		// "class=C234"
		path, ok := s.Find(".C234 a").Attr("href")
		// If no description url found, stop here
		if ok {
			// Get name from a "class=ttl" tag.
			// Remove dirty spaces before and after title.
			name := strings.TrimSpace(s.Find(".ttl").Text())
			// Build the real url
			url := baseURL + path
			// Store results
			t := Torrent{
				DescURL: url,
				Name:    name,
			}
			torrents = append(torrents, t)
		}
	})

	return torrents, nil
}

// Search takes a user search as a parameter and
// returns clean torrent information fetched from archive.org
func Search(in string) ([]Torrent, error) {

	// Build url
	url, err := buildURL(in)
	if err != nil {
		return nil, fmt.Errorf("error while building url: %v", err)
	}
	log.Printf("successfully built url: %s\n", url)

	// Fetch url
	resp, err := fetch(url)
	if err != nil {
		return nil, fmt.Errorf("error while fetching url: %v", err)
	}
	defer resp.Body.Close()
	log.Printf("successfully fetched html content\n")

	// Parse html response
	torrents, err := parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error while parsing results: %v", err)
	}

	return torrents, nil
}
