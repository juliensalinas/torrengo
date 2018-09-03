package tpb

import (
	"fmt"
	"io"

	"github.com/juliensalinas/torrengo/core"
)

func parseProxiesPage(r io.Reader) ([]string, error) {
	var urls []string

	urls = append(urls, "https://pirateproxy.gdn")
	urls = append(urls, "https://toto.gdn")

	return urls, nil
}

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
