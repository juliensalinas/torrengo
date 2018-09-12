package arc

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/juliensalinas/torrengo/core"
)

// parseDescPage parses the torrent description page and extracts the torrent file url
func parseDescPage(r io.Reader) (string, error) {
	// Load html response into GoQuery
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return "", fmt.Errorf("could not load html response into GoQuery: %v", err)
	}

	// Get the torrent file path from a "<a href=...>"" whose class starts with
	// "format-summary" and whose text contains the word "TORRENT"
	fileType := doc.Find(".format-summary ").First().Text()
	if fileType == "" {
		return "", fmt.Errorf("could not find the .format-summary tag on description page")
	}

	if !strings.Contains(fileType, "TORRENT") {
		return "", fmt.Errorf("could not find the TORRENT string in fileType")
	}

	path, ok := doc.Find(".format-summary ").First().Attr("href")
	if !ok {
		return "", fmt.Errorf("could not find a torrent file on the description page")
	}
	fileURL := baseURL + path

	return fileURL, nil
}

// FindAndDlFile opens the torrent description page and downloads the torrent
// file.
// A user timeout is set.
// Returns the local path of downloaded torrent file.
func FindAndDlFile(descURL string, timeout time.Duration) (string, error) {
	// Create an http client with user timeout
	client := &http.Client{
		Timeout: timeout,
	}

	// Fetch url
	resp, err := core.Fetch(descURL, client)
	if err != nil {
		return "", fmt.Errorf("error while fetching url: %v", err)
	}
	defer resp.Body.Close()

	// Parse html response
	fileURL, err := parseDescPage(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error while parsing torrent description page: %v", err)
	}

	// Download torrent
	filePath, err := core.DlFile(fileURL, client)
	if err != nil {
		return "", fmt.Errorf("error while downloading torrent file: %v", err)
	}

	return filePath, nil
}
