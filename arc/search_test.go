package arc

import (
	"testing"
	"time"
)

func TestLookup(t *testing.T) {
	torrents, err := Lookup("Monte Cristo", 30*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	if len(torrents) == 0 {
		t.Fatal("Found no torrent.")
	}

	if torrents[0].Name == "" {
		t.Fatal("Torrents have no name.")
	}
	if torrents[0].DescURL == "" {
		t.Fatal("Torrents have no Description URL.")
	}
}
