package ygg

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/juliensalinas/torrengo/core"
)

func parseDescPage(html string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", fmt.Errorf("could not load html response into GoQuery: %v", err)
	}

	// file url is located in the 1st <a> of the 2nd <td> of the first <tbody> of class infos-torrents
	fileURL, ok := doc.Find(".infos-torrent tbody").Eq(0).Find("tr td").Eq(1).Find("a").Eq(0).First().Attr("href")
	if !ok {
		return "", fmt.Errorf("could not find a torrent file on the description page")
	}

	return fileURL, nil
}

// FindAndDlFile authenticates user, opens the torrent description page,
// and downloads the torrent file.
// Returns the local path of downloaded torrent file.
// A user timeout is set.
func FindAndDlFile(descURL string, in string, userID string, userPass string,
	timeout time.Duration, client *http.Client) (string, error) {
	// Set timeout
	client.Timeout = timeout

	// Authenticate user and create http client that handles cookie and timeout
	client, err := authUser(userID, userPass, client)
	if err != nil {
		return "", fmt.Errorf("error while authenticating: %v", err)
	}
	decodedDescURL, err := url.ParseRequestURI(descURL)
	if err != nil {
		return "", fmt.Errorf("error while decoding url for cookie passing: %w", err)
	}
	cookies := client.Jar.Cookies(decodedDescURL)

	// Fetch url
	html, err := core.Fetch(context.TODO(), descURL, cookies)
	if err != nil {
		return "", fmt.Errorf("error while fetching url: %v", err)
	}

	// Parse html response
	filePath, err := parseDescPage(html)
	if err != nil {
		return "", fmt.Errorf("error while parsing torrent description page: %v", err)
	}

	fileURL := "https://" + baseURL + filePath
	filePathOnDisk, err := core.DlFile(fileURL, in, client)
	if err != nil {
		return "", fmt.Errorf("error while downloading torrent file: %v", err)
	}

	return filePathOnDisk, nil
}
