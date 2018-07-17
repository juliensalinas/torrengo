package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/onrik/logrus/filename"
	log "github.com/sirupsen/logrus"

	"github.com/juliensalinas/torrengo/arc"
	"github.com/juliensalinas/torrengo/td"
	"github.com/olekukonko/tablewriter"
)

// torrent contains meta information about the torrent
type torrent struct {
	fileURL string
	magnet  string
	// Description url containing more info about the torrent including the torrent file address
	descURL string
	name    string
	size    string
	// seeders and leechers could be int but tablewriter expects strings
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

func init() {
	// Log as JSON instead of the default ASCII formatter
	log.SetFormatter(&log.JSONFormatter{})

	// Only log the warning severity or above
	log.SetLevel(log.DebugLevel)

	// Log filename and line number.
	// Should be removed from production because adds a performance cost.
	log.AddHook(filename.NewHook())
}

// TODO: improve interaction with user
func main() {
	// Get command line flags and arguments
	websitePtr := flag.String("w", "all", "website you want to search: archive | torrentdownloads | all")
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
	case "torrentdownloads":
		tdTorrents, err := td.Lookup(s.in)
		if err != nil {
			log.Fatal(err)
		}
		for _, tdTorrent := range tdTorrents {
			t := torrent{
				descURL:  tdTorrent.DescURL,
				name:     tdTorrent.Name,
				size:     tdTorrent.Size,
				leechers: tdTorrent.Leechers,
				seeders:  tdTorrent.Seeders,
				source:   "TorrentDownloads",
			}
			s.out = append(s.out, t)
		}
	case "all":
		fmt.Println("Lookup all")
	}

	// Sort results (on seeders)
	// TODO: broken. Need to convert strings to int.
	s.sortOut()

	// Render the list of results to user in terminal
	render(s.out)

	// Read from user input the index of torrent we want to download
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Please select a torrent to download (enter its index): ")
	var index int
	for {
		indexStr, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Could not read your input, please try again (should be an integer):")
			continue
		}
		index, err = strconv.Atoi(strings.TrimSuffix(indexStr, "\n"))
		if err != nil {
			fmt.Println("Please enter an integer:")
			continue
		}
		break
	}

	var filePath string

	// Download torrent
	switch s.sourceToLookup {
	case "archive":
		filePath, err = arc.Download(s.out[index].descURL)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Here is your torrent file: %s\n", filePath)
	case "torrentdownloads":
		td.Download(s.out[index].descURL)
	case "all":
		fmt.Println("Download all")
	}

	// Open torrent in client
	switch s.sourceToLookup {
	case "archive":
		log.Debug("open %s with torrent client.", filePath)
		log.WithFields(log.Fields{
			"filePath": filePath,
			"client":   "Deluge",
		}).Debug("Opening file with torrent client")
		fmt.Println("Opening torrent in client...")
		cmd := exec.Command("deluge", filePath)
		// Use Start() instead of Run() because do not want to wait for the torrent
		// client process to complete (detached process).
		err := cmd.Start()
		if err != nil {
			log.Fatalf("Could not open your torrent in client, you need to do it manually: %s\n", err)
		}
	case "torrentdownloads":
	case "all":
		fmt.Println("Open all")

	}

}
