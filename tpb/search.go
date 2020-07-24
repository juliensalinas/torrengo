// Package tpb searches and extracts magnet link from ThePirateBay
//
// No check is done here regarding the user input. This check should be
// achieved by the caller.
// Parsing is achieved thanks to the GoQuery library.
// Comments common to all scraping libs are already done in the arc package which is very
// similar to this package. Only additional comments specific to this lib are present here.
//
// Torrent search is achieved by Lookup(). All useful information is located in the search result page.
// No need to open a second page.
// Input is a search string.
// Output is a slice of maps made up of the following keys:
//
// - Magnet: the torrent magnet
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
package tpb

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/juliensalinas/torrengo/core"
	log "github.com/sirupsen/logrus"
)

const proxiesListURL = "https://pirateproxy.wtf"

// Torrent contains meta information about the torrent
type Torrent struct {
	Magnet  string
	Name    string
	Size    string
	UplDate string
	// Seeders and Leechers are converted to -1 if cannot be converted to integers
	Seeders  int
	Leechers int
}

// A typical final url looks like:
// baseURL + /search/dumas/0/99/0
func buildSearchURL(baseURL string, in string) (string, error) {
	var URL *url.URL
	URL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("error during url parsing: %v", err)
	}

	URL.Path += "/search.php"
	q := URL.Query()
	q.Set("q", in)
	URL.RawQuery = q.Encode()

	return URL.String(), nil
}

func parseSearchPage(html string) ([]Torrent, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("could not load html response into GoQuery: %v", err)
	}

	// torrents stores a list of torrents made up of the torrent description url,
	// its name, its size, its seeders, and its leechers
	var torrents []Torrent

	// Results are located in a clean list
	doc.Find("#torrents li").Each(func(i int, s *goquery.Selection) {
		var t Torrent
		// Magnet is the href of the 4th <a> tag
		magnet, ok := s.Find("span").Eq(3).Find("a").First().Attr("href")
		if !ok {
			log.Debug("Could not find a magnet for a torrent so ignoring it")
			return
		}
		t.Magnet = magnet

		// Torrent name is the text of the <a> tag in the 2nd <span>
		t.Name = s.Find("span").Eq(1).Find("a").First().Text()

		// Upload date, size, seeders, and leechers, are the text of
		// other <span> tags.
		t.UplDate = s.Find("span").Eq(2).Text()
		t.Size = s.Find("span").Eq(4).Text()

		// We convert seeders and leechers to integers and
		// conversion fails we convert it to -1.
		seedersStr := s.Find("span").Eq(5).Text()
		seedersStr = strings.TrimSpace(seedersStr)
		seeders, err := strconv.Atoi(seedersStr)

		if err != nil {
			seeders = -1
		}
		t.Seeders = seeders

		leechersStr := s.Find("span").Eq(6).Text()
		leechersStr = strings.TrimSpace(leechersStr)
		leechers, err := strconv.Atoi(leechersStr)
		if err != nil {
			leechers = -1
		}
		t.Leechers = leechers

		torrents = append(torrents, t)
	})

	return torrents, nil
}

// checkEmptyResp checks whether the tpb response contains the
// #searchResult id, otherwise it means the site is broken
func checkEmptyResp(html string) bool {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return false
	}

	if doc.Find("#torrents").Nodes == nil {
		return false
	}

	return true
}

// Lookup takes a user search as a parameter and
// returns clean torrent information fetched from ThePirateBay.
// It first looks for the ThePirateBay proxies and then
// concurrently fetches all of them and retrieve results from
// the quickest one after checking that the latter is not broken.
// A custom user timeout is set.
func Lookup(in string, timeout time.Duration) ([]Torrent, error) {
	client := &http.Client{
		Timeout: timeout,
	}

	// Retrieve tpb proxies urls
	proxiesList, err := getProxies(client)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving proxies: %v", err)
	}

	// Create channels for communicating http response and termination
	// event in case of error
	htmlCh := make(chan string)
	htmlErrCh := make(chan struct{})

	// For each tpb proxy, launch the same request through a new
	// goroutine
	for _, baseURL := range proxiesList {
		fullURL, err := buildSearchURL(baseURL, in)
		if err != nil {
			log.WithFields(log.Fields{
				"err":     err,
				"baseURL": baseURL,
			}).Info("Could not build url for one of the TPB proxies")
			continue
		}
		go func(url string, localTimeout time.Duration) {
			// localClient := &http.Client{
			// 	Timeout: localTimeout,
			// }
			html, err := core.Fetch(context.TODO(), url)
			if err != nil {
				log.WithFields(log.Fields{
					"url": url,
				}).Debug("Broken proxy")
				htmlErrCh <- struct{}{}
				return
			}

			ok := checkEmptyResp(html)
			if !ok {
				log.WithFields(log.Fields{
					"url": url,
				}).Debug("Broken proxy (code 200 but empty response)")
				htmlErrCh <- struct{}{}
				return
			}
			log.WithFields(log.Fields{
				"url": url,
			}).Debug("Found a working proxy")

			htmlCh <- html
		}(fullURL, timeout)

	}

	var torrents []Torrent

	// From goroutines receive termination event (in case of error) or
	// http response. If http response received, it means the tpb proxy
	// worked properly and was the fastest to answer so parse results from html page
	// and leave.
	//
	// TODO(juliensalinas): once we have a working proxy, close all the other
	// pending goroutines.
	for i := 0; i < len(proxiesList); i++ {
		select {
		case <-htmlErrCh:
		case html := <-htmlCh:
			torrents, err = parseSearchPage(html)
			if err != nil {
				return nil, fmt.Errorf("error while parsing torrent search results: %v", err)
			}

			return torrents, nil
		}
	}

	return nil, fmt.Errorf("no tpb proxy working")
}
