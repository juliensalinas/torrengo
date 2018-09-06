package tpb

import (
	"fmt"
	"io"

	"github.com/PuerkitoBio/goquery"
	"github.com/juliensalinas/torrengo/core"
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

		// TPB site url is the href of a tag whose class is "site"
		s.Find(".site a ").Each(func(i int, ss *goquery.Selection) {
			u, ok := ss.Attr("href")
			if ok {
				url = u
				urls = append(urls, url)
			}
		})
	})

	return urls, nil
}

// getProxies returns a list of all tpb urls
func getProxies() ([]string, error) {
	resp, err := core.Fetch(proxiesListURL)
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
