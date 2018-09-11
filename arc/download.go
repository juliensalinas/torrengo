package arc

import (
	"fmt"
	"io"
	"strings"

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

	var fileURL string

	doc.Find(".format-summary ").Each(func(i int, s *goquery.Selection) {
		// Get the torrent file path from a "<a href=...>"" whose class starts with
		// "format-summary" and whose text contains the word "TORRENT"
		fileType := s.Text()
		if strings.Contains(fileType, "TORRENT") {
			path, ok := s.Attr("href")
			if ok {
				fileURL = baseURL + path
				return
			}
		}
	})
	if fileURL == "" {
		return "", fmt.Errorf("could not find a torrent file on the description page")
	}

	return fileURL, nil
}

// FindAndDlFile opens the torrent description page and downloads the torrent
// file. Returns the local path of downloaded torrent file.
func FindAndDlFile(descURL string) (string, error) {
	// Fetch url
	resp, err := core.Fetch(descURL, nil)
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
	filePath, err := core.DlFile(fileURL, nil)
	if err != nil {
		return "", fmt.Errorf("error while downloading torrent file: %v", err)
	}

	return filePath, nil
}
