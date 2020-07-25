package core

import (
	"context"
	"fmt"
	"io"
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

// Fetch opens a url with a custom context passed by the caller.
// It uses ChromeDP under the hood in order to emulate a real browser
// running on Pixel 2 XL, and thus properly handle Javascript.
func Fetch(ctx context.Context, url string, cookies []*http.Cookie) (string, error) {
	var html string

	chromedpCTX, cancel := chromedp.NewContext(ctx)
	defer cancel()

	// TODO(juliensalinas): check status code of the response, but not so easy
	// with the current version of ChromeDP. A new version should fix this:
	// https://github.com/chromedp/chromedp/issues/105
	err := chromedp.Run(chromedpCTX,
		setCookies(cookies),
		chromedp.Emulate(device.Pixel2XL),
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(chromedpCTX context.Context) error {
			node, err := dom.GetDocument().Do(chromedpCTX)
			if err != nil {
				return err
			}
			html, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(chromedpCTX)
			return err
		}),
	)

	if err != nil {
		return "", fmt.Errorf("could not download page: %w", err)
	}

	return html, nil
}

// DlFile downloads the torrent with a custom client created by user and returns the path of
// downloaded file.
// The name of the downloaded file is made up of the search arguments + the
// Unix timestamp to avoid collision. Ex: comte_de_montecristo_1581064034469619222.torrent
func DlFile(fileURL string, in string, client *http.Client) (string, error) {
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

func setCookies(cookies []*http.Cookie) chromedp.Action {
	for _, cookie := range cookies {
		return chromedp.ActionFunc(func(ctx context.Context) error {
			expr := cdp.TimeSinceEpoch(time.Now().Add(180 * 24 * time.Hour))
			success, err := network.SetCookie(cookie.Name, cookie.Value).
				WithExpires(&expr).
				WithDomain("localhost").
				WithHTTPOnly(true).
				Do(ctx)
			if err != nil {
				return err
			}
			if !success {
				return fmt.Errorf("could not set cookie %q to %q", cookie.Name, cookie.Value)
			}
			return nil
		})
	}

	// TODO(juliensalinas): doesn't work. Need to understand how to return a proper
	// empty chromedp.Action.
	return nil
}

// func SetCookie(name, value, domain, path string, httpOnly, secure bool) chromedp.Action {
// 	return chromedp.ActionFunc(func(ctx context.Context) error {
// 		expr := cdp.TimeSinceEpoch(time.Now().Add(180 * 24 * time.Hour))
// 		success, err := network.SetCookie(name, value).
// 			WithExpires(&expr).
// 			WithDomain(domain).
// 			WithPath(path).
// 			WithHTTPOnly(httpOnly).
// 			WithSecure(secure).
// 			Do(ctx)
// 		if err != nil {
// 			return err
// 		}
// 		if !success {
// 			return fmt.Errorf("could not set cookie %s", name)
// 		}
// 		return nil
// 	})
// }
