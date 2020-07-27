package ygg

import (
	"testing"
	"time"
)

func TestFindAndDlFile(t *testing.T) {
	id := ""
	pass := ""

	_, client, err := Lookup("Monte Cristo", 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	url, err := FindAndDlFile("https://www2.yggtorrent.si/torrent/ebook/audio/297687-alexandre+dumas+-+le+comte+de+monte-cristo+tome+1+2015+mp3+128kbps",
		"Monte Cristo", id, pass, 10*time.Second, client)

	if err != nil {
		t.Fatal(err)
	}

	if url == "" {
		t.Fatal("Got an empty url")
	}
}
