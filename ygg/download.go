package ygg

import (
	"fmt"
	"io"

	"github.com/PuerkitoBio/goquery"
	"github.com/juliensalinas/torrengo/core"
)

func parseDescPage(r io.Reader) (string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return "", fmt.Errorf("could not load html response into GoQuery: %v", err)
	}

	var fileURL string
	var fileURLIsOk bool

	doc.Find(".infos-torrent tbody").Eq(1).Find("tr td").Eq(1).Find("a").Eq(0).Each(func(i int, s *goquery.Selection) {
		fmt.Println(i)
		fileURL, fileURLIsOk = s.Attr("href")
	})
	if !fileURLIsOk {
		return "", fmt.Errorf("could not find a torrent file on the description page")
	}

	return fileURL, nil
}

// FindAndDlFile authenticates user, opens the torrent description page,
// and and downloads the torrent file.
// Returns the local path of downloaded torrent file.
func FindAndDlFile(descURL string, userID string, userPass string) (string, error) {
	// Authenticate user
	httpClient, err := authUser(userID, userPass)
	if err != nil {
		return "", fmt.Errorf("error while authenticating: %v", err)
	}

	// Fetch url
	resp, err := core.Fetch(descURL, httpClient)
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
	filePath, err := core.DlFile(fileURL, httpClient)
	if err != nil {
		return "", fmt.Errorf("error while downloading torrent file: %v", err)
	}

	return filePath, nil
}
