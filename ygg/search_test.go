package ygg

import (
	"testing"
	"time"
)

func TestLookup(t *testing.T) {
	torrents, _, err := Lookup("Monte cristo", 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	if len(torrents) == 0 {
		t.Fatal("Not torrent found.")
	}
}
