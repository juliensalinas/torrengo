package tpb

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/juliensalinas/torrengo/core"
	log "github.com/sirupsen/logrus"
)

// parseProxiesPage retrieves all the tpb urls from the html page
func parseProxiesPage(html string) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("could not load html response into GoQuery: %v", err)
	}

	// urls stores a list of tpb potential sites
	var urls []string

	// Results are located in a clean html <table>
	doc.Find(".proxies tbody tr").Each(func(i int, s *goquery.Selection) {
		// TPB site url is the href of a tag whose class is "site"
		url := strings.ToLower(s.Find("a").First().Text())
		if url == "" {
			log.Debug("could not find an url for a proxy")
			return
		}
		urls = append(urls, "https://"+url)
	})

	return urls, nil
}

// getProxies returns a list of all tpb urls
func getProxies(client *http.Client) ([]string, error) {
	html, err := core.Fetch(context.TODO(), proxiesListURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error while fetching url: %v", err)
	}

	urls, err := parseProxiesPage(html)
	if err != nil {
		return nil, fmt.Errorf("error while parsing torrent search results: %v", err)
	}

	return urls, nil
}
