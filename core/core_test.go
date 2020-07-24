package core

import (
	"context"
	"strings"
	"testing"
)

func TestFetch(t *testing.T) {
	// Testing a site that is supposed to use the Cloudflare challenge.
	// (Checking your browser before accessing xxx).
	html, err := Fetch(context.Background(), "https://support.litebit.eu/hc/en-us")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(html, "CloudFlare") {
		t.Fatal("Website triggered a Cloudflare challenge while it shouldn't have.")
	}
}
