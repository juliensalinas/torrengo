package td

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	cfScraper "github.com/juliensalinas/go-cloudflare-scraper"
	"github.com/juliensalinas/torrengo/core"
	log "github.com/sirupsen/logrus"
)

// parseDescPage parses the torrent description page and extracts the torrent file url
// + the magnet link
func parseDescPage(r io.Reader) (fileURL string, magnet string, err error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return "", "", fmt.Errorf("could not load html response into GoQuery: %v", err)
	}

	// Get the torrent url from a tag containing an image whose alt attribute is
	// "Download torrent"
	fileURL, fileURLOk := doc.Find("img[alt='Download torrent']").First().Parent().Attr("href")
	if !fileURLOk {
		log.Debug("Could not find a file URL")
	}

	// Get the magnet link from an <a> tag containing a "Magnet Link" text
	magnet, magnetOk := doc.Find("a:contains('Magnet Link')").First().Attr("href")
	if !magnetOk {
		log.Debug("Could not find a magnet link")
	}

	if !fileURLOk && !magnetOk {
		return "", "", fmt.Errorf("could not find neither a torrent file nor a magnet link on the description page")
	}

	return fileURL, magnet, nil
}

// DlFileFromCloudflare downloads the torrent file.
// The file is protected by Cloudflare so need to use a dedicated lib to bypass
// it. Does not work 100% of the time...
func DlFileFromCloudflare(fileURL string, timeout time.Duration) (string, error) {
	// Get torrent file name from url
	s := strings.Split(fileURL, "/")
	fileName := s[len(s)-1]
	s = strings.Split(fileName, ".torrent")
	fileName = s[0]
	fileName += ".torrent"

	// Create local torrent file
	out, err := os.Create(fileName)
	if err != nil {
		return "", fmt.Errorf("could not create the torrent file named %s: %v", fileName, err)
	}
	defer out.Close()

	// Initialize the Cloudflare scraping lib (also automatically adds a user-agent)
	cfScraper, err := cfScraper.NewTransport(http.DefaultTransport)
	if err != nil {
		return "", fmt.Errorf("could not initialize Cloudflare scraper: %v", err)
	}
	client := http.Client{
		Transport: cfScraper,
		Timeout:   timeout,
	}

	// Download torrent
	resp, err := client.Get(fileURL)
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

// ExtractTorAndMag opens the torrent description page and extracts the torrent
// file url + the magnet link.
// A user timeout is set.
func ExtractTorAndMag(descURL string, timeout time.Duration) (fileURL string, magnet string, err error) {
	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := core.Fetch(descURL, client)
	if err != nil {
		return "", "", fmt.Errorf("error while fetching url: %v", err)
	}
	defer resp.Body.Close()

	fileURL, magnet, err = parseDescPage(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("error while parsing torrent description page: %v", err)
	}

	return fileURL, magnet, nil
}
