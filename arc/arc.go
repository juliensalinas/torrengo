// Package arc searches archive.org and returns a clean map of torrents found
// on the first page.
// No check is done here regarding the user input. This check should be
// achieved by the caller.
package arc

import (
	"io/ioutil"
	"log"
	// "fmt"
	"net/http"
	"net/url"
)

const baseURL string = "https://archive.org"
const userAgent string = "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.62 Safari/537.36"

// buildURL encodes the user search keywords into a proper url.
// A typical final url looks like:
// https://archive.org/search.php?query=Dumas%20AND%20format%3A%22Archive%20BitTorrent%22
func buildURL(in string) (string, error) {
	// Add the following suffix to the query in order for archive.org
	// to return torrents only.
	in += ` AND format:"Archive BitTorrent"`

	// Encode baseURL as an url.URL type (Parse expects a pointer)
	// so we can work on it more easily.
	var URL *url.URL
	URL, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	// Create base path of URL.
	URL.Path += "/search.php"

	// Add GET parameters.
	params := url.Values{}
	params.Add("query", in)
	URL.RawQuery = params.Encode()

	return URL.String(), nil
}

// fetch opens a url and returns the resulting html page.
// Cannot use the straight http.Get function because need to
// modify headers in order to set a fake user-agent.
func fetch(url string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Set the fake user agent.
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// parse parses an html slice of bytes and returns a clean list
// of torrents found in this page.
func parse(resp []byte) ([][]string, error) {

	return nil, nil

}

// Search takes a user search as a parameter and
// returns clean torrent information fetched from archive.org
func Search(in string) ([][]string, error) {

	url, err := buildURL(in)
	if err != nil {
		return nil, err
	}
	log.Printf("Successfully built url: %s\n", url)

	resp, err := fetch(url)
	if err != nil {
		return nil, err
	}
	log.Printf("Successfully fetched the following content: \n%s\n", resp)

	torrents, err := parse(resp)
	if err != nil {
		return nil, err
	}

	return torrents, nil
}
