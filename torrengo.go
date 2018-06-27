package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/juliensalinas/torrengo/arc"
	"github.com/olekukonko/tablewriter"
)

// torrent stores all the final torrent details
type torrent struct {
	fileURL string
	magnet  string
	// Description url containing more info about the torrent including the torrent file address
	descURL  string
	name     string
	size     string
	seeders  string
	leechers string
	// Date of upload
	uplDate string
	// Website the torrent is coming from
	source string
}

// search represents the user search
type search struct {
	in             string
	out            []torrent
	sourceToLookup string
}

// cleanIn cleans the user search input
func (s *search) cleanIn() error {
	// Clean user input by removing useless spaces
	strings.TrimSpace(s.in)

	// If user input is empty raise an error
	if s.in == "" {
		return fmt.Errorf("user input should not be empty")
	}

	return nil
}

// sortOut sorts torrents list based on number of seeders (top down)
func (s *search) sortOut() {
	sort.Slice(s.out, func(i, j int) bool {
		return s.out[i].seeders > s.out[j].seeders
	})
}

// render renders torrents in a tabular user-friendly way with colors in terminal
func render(torrents []torrent) {
	// Turn type []torrent to type [][]string because this is what tablewriter expects
	var renderedTorrents [][]string
	for i, t := range torrents {
		renderedTorrent := []string{
			strconv.Itoa(i),
			t.name,
			t.size,
			t.seeders,
			t.leechers,
			t.uplDate,
			t.source,
		}
		renderedTorrents = append(renderedTorrents, renderedTorrent)
	}

	// Render results using tablewriter
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Index", "Name", "Size", "Seeders", "Leechers", "Date of upload", "Source"})
	table.SetRowLine(true)
	table.SetColumnColor(
		tablewriter.Colors{tablewriter.Normal, tablewriter.Normal},
		tablewriter.Colors{tablewriter.Normal, tablewriter.Normal},
		tablewriter.Colors{tablewriter.Normal, tablewriter.Normal},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiGreenColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiRedColor},
		tablewriter.Colors{tablewriter.Normal, tablewriter.Normal},
		tablewriter.Colors{tablewriter.Normal, tablewriter.Normal},
	)
	table.AppendBulk(renderedTorrents)
	table.Render()
}

func main() {
	// Show line number during logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Get command line flags and arguments
	websitePtr := flag.String("w", "all", "website you want to search: archive | all")
	flag.Parse()
	args := flag.Args()

	// If no command line argument is supplied, then we stop here
	if len(args) == 0 {
		os.Exit(1)
	}

	// Initialize the user search with the user input and sourceToLookup, and out is zeroed.
	// Concatenate all input arguments into one single string in case user does not use quotes.
	s := search{
		in:             strings.Join(args, " "),
		sourceToLookup: *websitePtr,
	}

	// Clean user input
	err := s.cleanIn()
	if err != nil {
		log.Fatal(err)
	}

	// Search torrents
	switch s.sourceToLookup {
	case "archive":
		arcTorrents, err := arc.Lookup(s.in)
		if err != nil {
			log.Fatal(err)
		}
		for _, arcTorrent := range arcTorrents {
			t := torrent{
				descURL: arcTorrent.DescURL,
				name:    arcTorrent.Name,
				source:  "Archive",
			}
			s.out = append(s.out, t)
		}
	case "all":
		fmt.Println("all")
	}

	// Sort results (on seeders)
	s.sortOut()

	// Render the list of results to user in terminal
	render(s.out)

}
