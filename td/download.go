package td

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/juliensalinas/torrengo/core"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/publicsuffix"
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

// DlFile downloads the torrent
// file.
// A user timeout is set.
// Returns the local path of downloaded torrent file.
func DlFile(fileURL string, timeout time.Duration) (string, error) {
	// Create an http client with user timeout.
	// Init cookies.
	// Using the publicsuffix list is recommended by Go docs.
	cookieJar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	client := &http.Client{
		Timeout: timeout,
		Jar:     cookieJar,
	}

	// Turn fileURL into proper url object
	urlObj, err := url.Parse(fileURL)
	if err != nil {
		return "", fmt.Errorf("could not turn descURL into proper url object: %v", err)
	}

	// Validate Cloudflare
	client, err = core.BypassCloudflare(*urlObj, client)

	// Fetch url
	resp, err := core.Fetch(fileURL, client)
	if err != nil {
		return "", fmt.Errorf("error while fetching url: %v", err)
	}
	defer resp.Body.Close()

	// Download torrent
	filePath, err := core.DlFile(fileURL, client)
	if err != nil {
		return "", fmt.Errorf("error while downloading torrent file: %v", err)
	}

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
