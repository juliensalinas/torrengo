package tpb

import (
	"context"
	"testing"
)

func TestGetProxies(t *testing.T) {
	urls, err := getProxies(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	want := 10
	if len(urls) < want {
		t.Fatalf("Got %v TPB urls, want at least %v", len(urls), want)
	}
}
