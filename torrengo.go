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

type torrent struct {
	fileURL  string
	magnet   string
	descURL  string // description url containing more info about the torrent including the torrent file address.
	name     string
	size     string
	seeders  string
	leechers string
	uplDate  string // date of upload.
	source   string // website the torrent is coming from.
}

func clean(in string) (string, error) {
	// Clean user input by removing useless spaces.
	clIn := strings.TrimSpace(in)

	// If user input is empty raise an error.
	if clIn == "" {
		return "", fmt.Errorf("user input should not be empty")
	}

	return clIn, nil
}

// render renders torrents in a tabular user-friendly way with colors in terminal.
func render(torrents []torrent) {
	// Turn type []torrent to type [][]string because this is what tablewriter expects.
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

	// Render results using tablewriter.
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
	// Show line number during logging.
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Get command line flags and arguments.
	websitePtr := flag.String("w", "all", "website you want to search: archive | all")
	flag.Parse()
	args := flag.Args()

	// If no command line argument is supplied, then we stop here.
	if len(args) == 0 {
		os.Exit(1)
	}

	// Concatenate all arguments into one single string in case user does not use quotes.
	in := strings.Join(args, " ")

	// Clean user input.
	clIn, err := clean(in)
	if err != nil {
		log.Fatal(err)
	}

	var torrents []torrent

	// Search torrents.
	switch *websitePtr {
	case "archive":
		arcTorrents, err := arc.Search(clIn)
		if err != nil {
			log.Fatal(err)
		}
		for _, arcTorrent := range arcTorrents {
			t := torrent{
				descURL: arcTorrent.DescURL,
				name:    arcTorrent.Name,
				source:  "Archive",
			}
			torrents = append(torrents, t)
		}
	case "all":
		fmt.Println("all")
	}

	// Sort torrents based on number of seeders (top down).
	sort.Slice(torrents, func(i, j int) bool {
		return torrents[i].seeders > torrents[j].seeders
	})

	// Render the list of results to user in terminal.
	render(torrents)

}
