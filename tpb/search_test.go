package tpb

import (
	"testing"
	"time"
)

func TestLookup(t *testing.T) {
	torrents, err := Lookup("Monte Cristo", 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	if len(torrents) == 0 {
		t.Fatal("Found no torrent.")
	}
}
