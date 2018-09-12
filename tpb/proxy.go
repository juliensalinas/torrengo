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
		// TPB site url is the href of a tag whose class is "site"
		url, ok := s.Find(".site a ").First().Attr("href")
		if !ok {
			log.Debug("could not find an url for a proxy")
			return
		}
		urls = append(urls, url)
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
