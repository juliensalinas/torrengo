package ygg

import (
	"testing"
	"time"
)

func TestLookup(t *testing.T) {
	_, _, err := Lookup("Le comte de Montecristo", time.Second)
	if err != nil {
		t.Fatalf("Cannot search torrent: %v", err)
	}
}
