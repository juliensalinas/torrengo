package td

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	cfScraper "github.com/juliensalinas/go-cloudflare-scraper"
	log "github.com/sirupsen/logrus"
)

// parseDescPage parses the torrent description page and extracts the torrent file url
// + the magnet link
func parseDescPage(r io.Reader) (fileURL string, magnet string, err error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return "", "", fmt.Errorf("could not load html response into GoQuery: %v", err)
	}

	doc.Find("img[alt='Download torrent']").Each(func(i int, s *goquery.Selection) {
		// Get the torrent url from a tag containing an image whose alt attribute is
		// "Download torrent"
		fileURL, _ = s.Parent().Attr("href")
	})
	doc.Find("a:contains('Magnet Link')").Each(func(i int, s *goquery.Selection) {
		// Get the magnet link from an <a> tag containing a "Magnet Link" text
		magnet, _ = s.Attr("href")
	})

	if fileURL == "" && magnet == "" {
		return "", "", fmt.Errorf("could not find neither a torrent file nor a magnet link on the description page")
	}

	return fileURL, magnet, nil
}

// DlFile downloads the torrent file.
// The file is protected by Cloudflare so need to use a dedicated lib to bypass
// it. Does not work 100% of the time...
func DlFile(fileURL string) (string, error) {
	// Get torrent file name from url
	s := strings.Split(fileURL, "/")
	fileName := s[len(s)-1]

	// Create local torrent file
	out, err := os.Create(fileName)
	if err != nil {
		return "", fmt.Errorf("could not create the torrent file named %s: %v", fileName, err)
	}
	defer out.Close()

	// Initialize the Cloudflare scraping lib
	cfScraper, err := cfScraper.NewTransport(http.DefaultTransport)
	if err != nil {
		return "", fmt.Errorf("could not initialize Cloudflare scraper: %v", err)
	}
	client := http.Client{Transport: cfScraper}

	// Download torrent
	fmt.Println(fileURL)
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
// file url + the magnet link
func ExtractTorAndMag(descURL string) (fileURL string, magnet string, err error) {
	resp, err := fetch(descURL)
	if err != nil {
		return "", "", fmt.Errorf("error while fetching url: %v", err)
	}
	defer resp.Body.Close()
	log.Debug("Successfully fetched html content.")

	fileURL, magnet, err = parseDescPage(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("error while parsing torrent description page: %v", err)
	}
	switch {
	case fileURL == "" && magnet != "":
		log.WithFields(log.Fields{
			"torrentURL": fileURL,
		}).Debug("Could not find a torrent file but successfully fetched a magnet link on the description page")
	case fileURL != "" && magnet == "":
		log.WithFields(log.Fields{
			"magnetLink": magnet,
		}).Debug("Could not find a magnet link but successfully fetched a torrent file on the description page")
	default:
		log.WithFields(log.Fields{
			"torrentURL": fileURL,
			"magnetLink": magnet,
		}).Debug("Successfully fetched a torrent file and a magnet link on the description page")
	}

	return fileURL, magnet, nil
}
