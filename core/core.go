package core

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/device"
)

// UserAgent is a customer browser user agent used in every HTTP connections
const UserAgent string = "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.62 Safari/537.36"

var cookieExpiry = time.Now().Add(10 * time.Minute)

// DlFileWithoutChrome downloads the torrent with a custom client created by user and returns the path of
// downloaded file.
// The name of the downloaded file is made up of the search arguments + the
// Unix timestamp to avoid collision. Ex: comte_de_montecristo_1581064034469619222.torrent
func DlFileWithoutChrome(fileURL string, in string, client *http.Client) (string, error) {
	// Get torrent file name from url
	fileName := strings.Replace(in, " ", "_", -1)
	fileName += "_" + strconv.Itoa(int(time.Now().UnixNano())) + ".torrent"

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
	req.Header.Set("User-Agent", UserAgent)
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

// FetchWithoutChrome fetches a URL using Go http client under the hood
// instead of Chrome.
func FetchWithoutChrome(url string, client *http.Client) (string, *http.Client, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", nil, fmt.Errorf("could not create request: %v", err)
	}

	req.Header.Set("User-Agent", UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("could not launch request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return "", nil, fmt.Errorf("status code error: %v", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("can't read response body: %w", err)
	}

	return string(body), client, nil
}

// Fetch opens a url with custom context and cookies passed by the caller.
// It uses ChromeDP under the hood in order to emulate a real browser
// running on Pixel 2 XL, and thus properly handle Javascript.
func Fetch(ctx context.Context, url string, cookies []*http.Cookie) (string, []*http.Cookie, error) {
	var html string
	var newCDPCookies []*network.Cookie
	var newCookies []*http.Cookie

	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	// TODO(juliensalinas): check status code of the response
	err := chromedp.Run(ctx,
		setCookies(ctx, cookies),
		chromedp.Emulate(device.Pixel2XL),
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Retrieve HTML response.
			node, err := dom.GetDocument().Do(ctx)
			if err != nil {
				return err
			}
			html, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
			if err != nil {
				return err
			}

			// Retrieve response cookies.
			newCDPCookies, err = network.GetAllCookies().Do(ctx)
			if err != nil {
				return err
			}

			newCookies = convertCookies(newCDPCookies)

			return nil
		}),
	)

	if err != nil {
		return "", nil, fmt.Errorf("could not download page: %w", err)
	}

	return html, newCookies, nil
}

// convertCookies converts ChromeDP cookies to Go http cookies.
func convertCookies(cookies []*network.Cookie) []*http.Cookie {
	var newCookies []*http.Cookie

	for _, cookie := range cookies {
		newCookie := http.Cookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Path:     cookie.Path,
			Domain:   cookie.Domain,
			Expires:  cookieExpiry,
			Secure:   cookie.Secure,
			HttpOnly: cookie.HTTPOnly,
		}
		newCookies = append(newCookies, &newCookie)
	}

	return newCookies
}

// setCookies retrieves Go http cookies and sets ChromeDP out of it.
//
// TODO(juliensalinas): try again to use network.SetCookies. Last
// time it failed with "invalid parameter -32602 for some reason".
func setCookies(ctx context.Context, cookies []*http.Cookie) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		for _, cookie := range cookies {
			expiry := cdp.TimeSinceEpoch(cookieExpiry)
			success, err := network.SetCookie(cookie.Name, cookie.Value).
				WithExpires(&expiry).
				WithDomain(cookie.Domain).
				WithPath(cookie.Path).
				WithHTTPOnly(cookie.HttpOnly).
				WithSecure(cookie.Secure).
				Do(ctx)
			if err != nil {
				return err
			}
			if !success {
				return fmt.Errorf("could not set cookie %v to %v", cookie.Name, cookie.Value)
			}
		}

		// Check that cookies were properly set.
		cookiesInBrowser, err := network.GetAllCookies().Do(ctx)
		if err != nil {
			return err
		}
		if len(cookiesInBrowser) != len(cookies) {
			return fmt.Errorf("cookies not properly set")
		}

		return nil
	})
}
