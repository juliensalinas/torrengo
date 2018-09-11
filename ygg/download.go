package ygg

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

// dlFile downloads the torrent file
func dlFile(fileURL string, client *http.Client) (string, error) {
	if client == nil {
		client = &http.Client{}
	}

	// Get torrent file name from url
	s := strings.Split(fileURL, "/")
	fileName := s[len(s)-1]

	// Create local torrent file
	out, err := os.Create(fileName)
	if err != nil {
		return "", fmt.Errorf("could not create the torrent file named %s: %v", fileName, err)
	}
	defer out.Close()

	// Download torrent
	req, err := http.NewRequest("GET", fileURL, nil)
	if err != nil {
		return "", fmt.Errorf("could not create request: %v", err)
	}
	req.Header.Set("User-Agent", core.UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not download the torrent file: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return "", fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	// Save torrent to disk
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("could not save the torrent file to disk: %v", err)
	}

	// Get absolute file path of torrent
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", fmt.Errorf("could not retrieve current directory of saved filed: %v", err)
	}
	filePath := dir + "/" + fileName

	return filePath, nil
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
	filePath, err := dlFile(fileURL, httpClient)
	if err != nil {
		return "", fmt.Errorf("error while downloading torrent file: %v", err)
	}

	return filePath, nil
}
