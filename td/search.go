// Package td searches torrentdownloads.me and returns a clean list of torrents found
// on the first page based on a user search.
// No check is done here regarding the user input. This check should be
// achieved by the caller.
// Package td also downloads the torrent file located on a webpage provided by user (very tricky
// because torrentdownloads has a Cloudflare protection so  does not work 100% of the time)
// or retrieves the magnet link.
// Parsing is achieved thanks to the GoQuery library.
//
// Input passed to the Search() function is a search string.
// Output is a slice of maps made up of 2 keys:
// descUrl: the torrent description dedicated url
// name: the torrent name
//
// Comments common to all scraping libs are already done in the arc package which is very
// similar to this package. Only additional comments specific to this lib are present here.
package td

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

const baseURL string = "https://www.torrentdownloads.me"
const userAgent string = "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.62 Safari/537.36"

// Torrent contains meta information about the torrent
type Torrent struct {
	DescURL  string
	Name     string
	Size     string
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
		// Many elements in inner_container are junk (ads, empty stuffs,...) so
		// we only take elements between 10 and 2 before the end
		if i > 9 && i < l-2 {
			// Get path to torrent description page from a "<a>" tag located
			// inside a <p> tag
			path, ok := s.Find("p a").Attr("href")
			if ok {
				// Get name from the same place as path
				name := strings.TrimSpace(s.Find("p a").Text())
				url := baseURL + path
				t := Torrent{
					DescURL: url,
					Name:    name,
				}
				// Get leechers, seeders and size from the 3 first <span> tags
				s.Find("span").Each(func(i int, ss *goquery.Selection) {
					switch i {
					case 1:
						leechersStr := ss.Text()
						leechers, err := strconv.Atoi(leechersStr)
						if err != nil {
							log.Fatal(err)
						}
						t.Leechers = leechers

					case 2:
						seedersStr := ss.Text()
						seeders, err := strconv.Atoi(seedersStr)
						if err != nil {
							log.Fatal(err)
						}
						t.Seeders = seeders

					case 3:
						size := ss.Text()
						t.Size = size
					}
				})
				torrents = append(torrents, t)
			}
		}
	})

	fmt.Println(torrents)
	return torrents, nil
}

// Lookup takes a user search as a parameter and
// returns clean torrent information fetched from torrentdownloads.me
func Lookup(in string) ([]Torrent, error) {

	url, err := buildSearchURL(in)
	if err != nil {
		return nil, fmt.Errorf("error while building url: %v", err)
	}
	log.WithFields(log.Fields{
		"url": url,
	}).Debug("Successfully built url.")

	resp, err := fetch(url)
	if err != nil {
		return nil, fmt.Errorf("error while fetching url: %v", err)
	}
	defer resp.Body.Close()
	log.Debug("Successfully fetched html content.")

	torrents, err := parseSearchPage(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error while parsing torrent search results: %v", err)
	}

	return torrents, nil
}
