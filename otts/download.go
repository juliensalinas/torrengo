package otts

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/juliensalinas/torrengo/core"

	"github.com/PuerkitoBio/goquery"
)

// parseDescPage parses the torrent description page and extracts the magnet link
func parseDescPage(html string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", fmt.Errorf("could not load html response into GoQuery: %v", err)
	}

	magnet, ok := doc.Find(".torrent-detail-page li a").Eq(0).First().Attr("href")
	if !ok {
		return "", fmt.Errorf("could not extract magnet link")
	}

	return magnet, nil
}

// ExtractMag opens the torrent description page and extracts the magnet link.
// A user timeout is set.
func ExtractMag(descURL string, timeout time.Duration) (string, error) {
	// client := &http.Client{
	// 	Timeout: timeout,
	// }

	html, err := core.Fetch(context.TODO(), descURL)
	if err != nil {
		return "", fmt.Errorf("error while fetching url: %v", err)
	}

	magnet, err := parseDescPage(html)
	if err != nil {
		return "", fmt.Errorf("error while parsing torrent description page: %v", err)
	}

	return magnet, nil
}
