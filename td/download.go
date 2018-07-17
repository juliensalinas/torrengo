package td

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

// parseDescPage parses the torrent description page and extracts the torrent file url
func parseDescPage(r io.Reader) (string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return "", fmt.Errorf("could not load html response into GoQuery: %v", err)
	}

	var fileURL string

	doc.Find("img[alt='Download torrent']").Each(func(i int, s *goquery.Selection) {
		path, ok := s.Parent().Attr("href")
		if ok {
			fileURL = baseURL + path
			return
		}
	})

	if fileURL == "" {
		return "", fmt.Errorf("could not find a torrent file on the description page")
	}

	return fileURL, nil
}

// dlFile downloads the torrent file
func dlFile(fileURL string) (string, error) {
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
	resp, err := http.Get(fileURL)
	if err != nil {
		return "", fmt.Errorf("could not download the torrent file: %v", err)
	}
	defer resp.Body.Close()

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

// Download opens the torrent description page and downloads the torrent
// file. Returns the local path of downloaded torrent file.
func Download(descURL string) (string, error) {
	resp, err := fetch(descURL)
	if err != nil {
		return "", fmt.Errorf("error while fetching url: %v", err)
	}
	defer resp.Body.Close()
	log.Debug("Successfully fetched html content.")

	fileURL, err := parseDescPage(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error while parsing torrent description page: %v", err)
	}
	log.WithFields(log.Fields{
		"url": fileURL,
	}).Debug("Successfully fetched torrent file url.")

	// filePath, err := dlFile(fileURL)
	// if err != nil {
	// 	return "", fmt.Errorf("error while downloading torrent file: %v", err)
	// }
	// log.WithFields(log.Fields{
	// 	"filePath": filePath,
	// }).Debug("Successfully dowloaded torrent file.")

	// return filePath, nil

	return "", nil
}
