package otts

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/juliensalinas/torrengo/core"

	"github.com/PuerkitoBio/goquery"
)

// parseDescPage parses the torrent description page and extracts the magnet link
func parseDescPage(r io.Reader) (string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return "", fmt.Errorf("could not load html response into GoQuery: %v", err)
	}

	magnet, ok := doc.Find(".download-links-dontblock li a").Eq(0).First().Attr("href")
	if !ok {
		return "", fmt.Errorf("could not extract magnet link")
	}

	return magnet, nil
}

// ExtractMag opens the torrent description page and extracts the magnet link.
// A user timeout is set.
func ExtractMag(descURL string, timeout time.Duration) (string, error) {
	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := core.Fetch(descURL, client)
	if err != nil {
		return "", fmt.Errorf("error while fetching url: %v", err)
	}
	defer resp.Body.Close()

	magnet, err := parseDescPage(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error while parsing torrent description page: %v", err)
	}

	return magnet, nil
}
