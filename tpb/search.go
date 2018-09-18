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
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/juliensalinas/torrengo/core"
	log "github.com/sirupsen/logrus"
)

const proxiesListURL = "https://proxybay.bz/"

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

	URL.Path += "/search/" + in + "/0/99/0"

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
	l := doc.Find("#searchResult tbody tr").Size()

	// Results are located in a clean html <table>
	doc.Find("#searchResult tbody tr").Each(func(i int, s *goquery.Selection) {
		// Last line of the table is not relevant
		if i < l-1 {
			var t Torrent
			// Magnet is the href of the 4th <a> tag
			magnet, ok := s.Find("a").Eq(3).First().Attr("href")
			if !ok {
				log.Debug("Could not find a magnet for a torrent so ignoring it")
				return
			}
			t.Magnet = magnet

			// Torrent name is the text of a tag whose class is "detLink"
			t.Name = s.Find(".detLink").First().Text()

			// Size and upload date are concatenated in a string in a <font> tag.
			// Each piece of info is comma separated.
			// We then remove spaces and unneeded text.
			text := s.Find("font").First().Text()
			textSlc := strings.Split(text, ",")
			if len(textSlc) > 1 { // A security just in case
				t.UplDate = strings.TrimSpace(strings.Replace(textSlc[0], "Uploaded", "", -1))
				t.Size = strings.TrimSpace(strings.Replace(textSlc[1], "Size", "", -1))
			}
			// Seeders and leechers are located in the 3rd and 4th <td>.
			// We convert it to integers and if conversion fails we convert it to -1.
			seedersStr := s.Find("td").Eq(2).First().Text()
			seeders, err := strconv.Atoi(seedersStr)
			if err != nil {
				seeders = -1
			}
			t.Seeders = seeders

			leechersStr := s.Find("td").Eq(3).First().Text()
			leechers, err := strconv.Atoi(leechersStr)
			if err != nil {
				leechers = -1
			}
			t.Leechers = leechers

			torrents = append(torrents, t)
		}
	})

	return torrents, nil
}

// checkEmptyResp checks whether the tpb response contains the
// #searchResult id, otherwise it means the site is broken
func checkEmptyResp(r io.Reader) bool {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return false
	}

	if doc.Find("#searchResult").Nodes == nil {
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
	httpRespCh := make(chan *http.Response)
	httpRespErrCh := make(chan struct{})

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
			localClient := &http.Client{
				Timeout: localTimeout,
			}
			resp, err := core.Fetch(url, localClient)
			if err != nil {
				log.WithFields(log.Fields{
					"url": url,
				}).Debug("Broken proxy (no code 200)")
				httpRespErrCh <- struct{}{}
				return
			}
			// A resp.Body cannot be read twice, so need to first extract
			// extract bodyBytes and use ioutil.NopCloser on it each time we
			// want to read the body.
			// Could be optimized because here we read the body 3 times instead
			// of 2 (should retrieve bodyBytes whithin checkEmptyResp when it
			// is read for the first time).
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
			ok := checkEmptyResp(resp.Body)
			if !ok {
				log.WithFields(log.Fields{
					"url": url,
				}).Debug("Broken proxy (code 200 but empty response)")
				httpRespErrCh <- struct{}{}
				return
			}
			log.WithFields(log.Fields{
				"url": url,
			}).Debug("Found a working proxy")
			resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
			httpRespCh <- resp
		}(fullURL, timeout)

	}

	var torrents []Torrent

	// From goroutines receive termination event (in case of error) or
	// http response. If http response received, it means the tpb proxy
	// worked properly and was the fastest to answer so parse results from html page
	// and leave.
	for i := 0; i < len(proxiesList); i++ {
		select {
		case <-httpRespErrCh:
		case resp := <-httpRespCh:
			torrents, err = parseSearchPage(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("error while parsing torrent search results: %v", err)
			}
			resp.Body.Close()
			return torrents, nil
		}
	}

	return nil, fmt.Errorf("no tpb proxy working")
}
