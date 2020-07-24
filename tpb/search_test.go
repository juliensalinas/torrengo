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

	if torrents[0].Name == "" {
		t.Fatal("Torrents have no name.")
	}
	if torrents[0].Magnet == "" {
		t.Fatal("Torrents have no magnet.")
	}
	if torrents[0].Size == "" {
		t.Fatal("Torrents have no size.")
	}
	if torrents[0].UplDate == "" {
		t.Fatal("Torrents have no Upload date.")
	}
	if torrents[0].Leechers == -1 {
		t.Fatal("Torrents have no leachers.")
	}
	if torrents[0].Seeders == -1 {
		t.Fatal("Torrents have no seeders.")
	}
}
