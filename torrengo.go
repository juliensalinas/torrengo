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
	descURL  string
	name     string
	size     string
	seeders  int
	leechers int
	// Date of upload
	uplDate string
	// Website the torrent is coming from
	source string
	// Local path where torrent was saved
	filePath string
}

var ft torrent

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
		return fmt.Errorf("User input should not be empty")
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
		seedersStr := strconv.Itoa(t.seeders)
		leechersStr := strconv.Itoa(t.leechers)
		renderedTorrent := []string{
			strconv.Itoa(i),
			t.name,
			t.size,
			seedersStr,
			leechersStr,
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

// getAndShowMagnet retrieves and displays magnet to user
func getAndShowMagnet() {
	fmt.Printf("Here is your magnet link: %s\n", ft.magnet)
}

// getAndShowTorrent retrieves and displays torrent file to user
func getAndShowTorrent() {
	var err error
	switch ft.source {
	case "arc":
		ft.filePath, err = arc.FindAndDlFile(ft.descURL)
	case "td":
		ft.filePath, err = td.DlFile(ft.fileURL)
	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Here is your torrent file: %s\n", ft.filePath)
}

func openMagOrTorInClient(resource string) {
	// Open torrent in client
	log.Debug("open %s with torrent client.", resource)
	log.WithFields(log.Fields{
		"resource": resource,
		"client":   "Deluge",
	}).Debug("opening magnet link or torrent file with torrent client")
	fmt.Println("opening torrent in client...")
	cmd := exec.Command("deluge", resource)
	// Use Start() instead of Run() because do not want to wait for the torrent
	// client process to complete (detached process).
	err := cmd.Start()
	if err != nil {
		log.Fatalf("could not open your torrent in client, you need to do it manually: %s\n", err)
	}
}

// TODO: improve interaction with user
// TODO: see when to log and when to fmt
func main() {
	// Get command line flags and arguments
	websitePtr := flag.String("w", "all", "website you want to search: arc | td | all")
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
	case "arc":
		arcTorrents, err := arc.Lookup(s.in)
		if err != nil {
			log.Fatal(err)
		}
		for _, arcTorrent := range arcTorrents {
			t := torrent{
				descURL: arcTorrent.DescURL,
				name:    arcTorrent.Name,
				source:  "arc",
			}
			s.out = append(s.out, t)
		}
	case "td":
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
				source:   "td",
			}
			s.out = append(s.out, t)
		}
	case "all":
		arcTorListCh := make(chan []torrent)
		arcSearchErrCh := make(chan error)
		tdTorListCh := make(chan []torrent)
		tdSearchErrCh := make(chan error)

		var arcSearchErr, tdSearchErr error

		go func() {
			arcTorrents, err := arc.Lookup(s.in)
			if err != nil {
				arcSearchErrCh <- err
			}
			var torList []torrent
			for _, arcTorrent := range arcTorrents {
				t := torrent{
					descURL: arcTorrent.DescURL,
					name:    arcTorrent.Name,
					source:  "arc",
				}
				torList = append(torList, t)
			}
			arcTorListCh <- torList
		}()
		select {
		case arcSearchErr = <-arcSearchErrCh:
			log.Errorf("the arc search goroutine quit with an error: %v", err)
		case arcTorList := <-arcTorListCh:
			s.out = append(s.out, arcTorList...)
		}
		go func() {
			tdTorrents, err := td.Lookup(s.in)
			if err != nil {
				tdSearchErrCh <- err
			}
			var torList []torrent
			for _, tdTorrent := range tdTorrents {
				t := torrent{
					descURL:  tdTorrent.DescURL,
					name:     tdTorrent.Name,
					size:     tdTorrent.Size,
					leechers: tdTorrent.Leechers,
					seeders:  tdTorrent.Seeders,
					source:   "td",
				}
				torList = append(torList, t)
			}
			tdTorListCh <- torList
		}()
		select {
		case tdSearchErr = <-tdSearchErrCh:
			log.Errorf("the td search goroutine quit with an error: %v", err)
		case tdTorList := <-tdTorListCh:
			s.out = append(s.out, tdTorList...)
		}

		if arcSearchErr != nil && tdSearchErr != nil {
			log.Fatal("all searches return an error...")
		}
	}

	// Sort results (on seeders)
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

	ft = s.out[index]

	// Download torrent
	switch s.sourceToLookup {
	case "arc":
		getAndShowTorrent()
		openMagOrTorInClient(ft.filePath)
	case "td":
		ft.fileURL, ft.magnet, err = td.ExtractTorAndMag(ft.descURL)
		if err != nil {
			log.Fatal(err)
		}
		switch {
		case ft.fileURL == "" && ft.magnet != "":
			getAndShowMagnet()
		case ft.fileURL != "" && ft.magnet == "":
			getAndShowTorrent()
		default:
			// Ask user to choose between file download and magnet download
			reader := bufio.NewReader(os.Stdin)
			fmt.Println("We found a torrent file and a magnet link, which one would you like to download?" +
				"\n1) Magnet link\n2) Torrent file (careful: not working 100% of the time)")
			var choice int
			for {
				choiceStr, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println("Could not read your input, please enter your choice (1 or 2):")
					continue
				}
				choice, err = strconv.Atoi(strings.TrimSuffix(choiceStr, "\n"))
				if err != nil {
					fmt.Println("Please enter an integer:")
					continue
				}
				break
			}
			switch choice {
			case 1:
				getAndShowMagnet()
				openMagOrTorInClient(ft.magnet)
			case 2:
				getAndShowTorrent()
				openMagOrTorInClient(ft.filePath)
			}
		}
	case "all":
		fmt.Println("Download all")
	}

}
