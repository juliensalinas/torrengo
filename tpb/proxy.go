package tpb

import (
	"fmt"
	"io"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/juliensalinas/torrengo/core"
	log "github.com/sirupsen/logrus"
)

// parseProxiesPage retrieves all the tpb urls from the html page
func parseProxiesPage(r io.Reader) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("could not load html response into GoQuery: %v", err)
	}

	// urls stores a list of tpb potential sites
	var urls []string

	// Results are located in a clean html <table>
	doc.Find("#proxyList tbody tr").Each(func(i int, s *goquery.Selection) {
		var url string
		var urlIsOk bool

		// TPB site url is the href of a tag whose class is "site"
		s.Find(".site a ").Each(func(i int, ss *goquery.Selection) {
			url, urlIsOk = ss.Attr("href")
		})
		if urlIsOk {
			urls = append(urls, url)
		} else {
			log.Debug("could not find a url for a proxy")
		}
	})

	return urls, nil
}

// getProxies returns a list of all tpb urls
func getProxies(client *http.Client) ([]string, error) {
	resp, err := core.Fetch(proxiesListURL, client)
	if err != nil {
		return nil, fmt.Errorf("error while fetching url: %v", err)
	}
	defer resp.Body.Close()

	urls, err := parseProxiesPage(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error while parsing torrent search results: %v", err)
	}

	return urls, nil
}
