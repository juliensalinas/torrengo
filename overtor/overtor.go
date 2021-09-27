package overtor

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cretz/bine/tor"
	"golang.org/x/net/html"
)

// the new v3 onion url for piratebay
const pirateURL = "http://piratebayo3klnzokct3wt5yyxb2vpebbuyjl7m623iaxmqhsd52coid.onion"

func DoTor() (string, error) {
	// Start tor with default config (can set start conf's DebugWriter to os.Stdout for debug logs)
	fmt.Println("Starting tor and fetching title of https://check.torproject.org, please wait a few seconds...")
	t, err := tor.Start(nil, nil)
	if err != nil {
		return "", err
	}
	//	defer t.Close()
	// Wait at most a minute to start network and get
	dialCtx, dialCancel := context.WithTimeout(context.Background(), time.Minute)
	defer dialCancel()
	// Make connection
	dialer, err := t.Dialer(dialCtx, nil)
	if err != nil {
		return "", err
	}
	if err := areWeTor(dialer); err != nil {
		return "", err
	}

	info, err := t.Control.GetInfo("net/listeners/socks")
	torAddr := info[0].Val
	//t.StopProcessOnClose = true
	return torAddr, nil
}

func areWeTor(dialer *tor.Dialer) error {
	httpClient := &http.Client{Transport: &http.Transport{DialContext: dialer.DialContext}}
	// Get /
	resp, err := httpClient.Get("https://check.torproject.org")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Grab the <title>
	parsed, err := html.Parse(resp.Body)
	if err != nil {
		return err
	}
	fmt.Printf("Title: %v\n", getTitle(parsed))
	return nil
}

func getTitle(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "title" {
		var title bytes.Buffer
		if err := html.Render(&title, n.FirstChild); err != nil {
			panic(err)
		}
		return strings.TrimSpace(title.String())
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if title := getTitle(c); title != "" {
			return title
		}
	}
	return ""
}
