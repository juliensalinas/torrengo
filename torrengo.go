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
	in              string
	out             []torrent
	sourcesToLookup []string
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
		// Replace -1 by unknown because more user-friendly
		seedersStr := strconv.Itoa(t.seeders)
		if seedersStr == "-1" {
			seedersStr = "Unknown"
		}
		leechersStr := strconv.Itoa(t.leechers)
		if leechersStr == "-1" {
			leechersStr = "Unknown"
		}
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
		fmt.Println("Could not retrieve the torrent file (see logs for more details).")
		log.WithFields(log.Fields{
			"descURL": ft.descURL,
			"error":   err,
		}).Fatal("Could not retrieve the torrent file")
	}
	fmt.Printf("Here is your torrent file: %s\n", ft.filePath)
}

func openMagOrTorInClient(resource string) {
	// Open torrent in client
	log.Debug("Open %s with torrent client.", resource)
	log.WithFields(log.Fields{
		"resource": resource,
		"client":   "Deluge",
	}).Debug("Opening magnet link or torrent file with torrent client")
	fmt.Println("opening torrent in client...")
	cmd := exec.Command("deluge", resource)
	// Use Start() instead of Run() because do not want to wait for the torrent
	// client process to complete (detached process).
	err := cmd.Start()
	if err != nil {
		fmt.Println("Could not open your torrent in client, you need to do it manually (see logs for more details).")
		log.WithFields(log.Fields{
			"resource": resource,
			"client":   "Deluge",
			"error":    err,
		}).Fatal("Could not open torrent in client")
	}
}

// rmDuplicates removes duplicates from slice
func rmDuplicates(elements []string) []string {
	encountered := map[string]bool{}

	// Create a map of all unique elements.
	for v := range elements {
		encountered[elements[v]] = true
	}

	// Place all keys from the map into a slice.
	result := []string{}
	for key := range encountered {
		result = append(result, key)
	}
	return result
}

func init() {
	// TODO: log to file
	// TODO: mention log path in every user error message
	// TODO: log as JSON for production

	// Log as JSON instead of the default ASCII formatter
	// log.SetFormatter(&log.JSONFormatter{})
	log.SetFormatter(&log.TextFormatter{})

	// Only log the warning severity or above
	log.SetLevel(log.DebugLevel)

	// Log filename and line number.
	// Should be removed from production because adds a performance cost.
	log.AddHook(filename.NewHook())
}

// TODO: improve interaction with user
func main() {
	// Get command line flags and arguments
	sourcesPtr := flag.String("w", "all", "A comma separated list of websites "+
		"you want to search (e.g. arc,td,tbp). Choices: arc | td | all. "+
		"\"all\" searches all websites.")
	flag.Parse()
	args := flag.Args()

	// If no command line argument is supplied, then we stop here
	if len(args) == 0 {
		fmt.Println("Please enter proper arguments (-h for help).")
		os.Exit(1)
	}

	// Initialize the user search with the user input and sourcesToLookup, and out is zeroed.
	// Remove possible duplicates from user input.
	// In case user chooses "all" as a source, convert it to the proper source names.
	// Stop if a user source is unknown.
	// Concatenate all input arguments into one single string in case user does not use quotes.
	sourcesSlc := strings.Split(*sourcesPtr, ",")
	cleanedSourcesSlc := rmDuplicates(sourcesSlc)
	for _, source := range cleanedSourcesSlc {
		if source == "all" {
			cleanedSourcesSlc = []string{"arc", "td"}
			break
		}
		if source != "arc" && source != "td" {
			fmt.Printf("This website is not correct: %v\n", source)
			log.WithFields(log.Fields{
				"sourcesList": cleanedSourcesSlc,
				"wrongSource": source,
			}).Fatal("Unknown source in user sources list")
		}
	}
	s := search{
		in:              strings.Join(args, " "),
		sourcesToLookup: cleanedSourcesSlc,
	}

	// Clean user input
	err := s.cleanIn()
	if err != nil {
		fmt.Println("Could not process your input (see logs for more details).")
		log.WithFields(log.Fields{
			"input": s.in,
			"error": err,
		}).Fatal("Could not clean user input")
	}

	// Channels for results
	arcTorListCh := make(chan []torrent)
	tdTorListCh := make(chan []torrent)

	// Channels for errors
	arcSearchErrCh := make(chan error)
	tdSearchErrCh := make(chan error)

	// Launch all torrent search goroutines
	for _, source := range s.sourcesToLookup {
		switch source {
		// User wants to search archive.org
		case "arc":
			// Concurrently search archive.org
			go func() {
				arcTorrents, err := arc.Lookup(s.in)
				if err != nil {
					arcSearchErrCh <- err
				}
				var torList []torrent
				for _, arcTorrent := range arcTorrents {
					t := torrent{
						descURL:  arcTorrent.DescURL,
						name:     arcTorrent.Name,
						size:     "unknown",
						leechers: -1,
						seeders:  -1,
						source:   "arc",
					}
					torList = append(torList, t)
				}
				arcTorListCh <- torList
			}()

		// User wants to search torrentdownloads.me
		case "td":

			// Concurrently search torrentdownloads.me
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
		}
	}

	// Initialize search errors
	var tdSearchErr, arcSearchErr error

	// Gather all goroutines results
	for _, source := range s.sourcesToLookup {
		switch source {
		case "arc":
			// Get results or error from archive.org
			select {
			case arcSearchErr = <-arcSearchErrCh:
				fmt.Println("An error occured during search on Archive.org.")
				log.WithFields(log.Fields{
					"input": s.in,
					"error": err,
				}).Error("The arc search goroutine broke")
			case arcTorList := <-arcTorListCh:
				s.out = append(s.out, arcTorList...)
			}
		case "td":
			// Get results or error from torrentdownloads.com
			select {
			case tdSearchErr = <-tdSearchErrCh:
				fmt.Println("An error occured during search on TorrentDownloads.com.")
				log.WithFields(log.Fields{
					"input": s.in,
					"error": err,
				}).Error("The td search goroutine broke")
			case tdTorList := <-tdTorListCh:
				s.out = append(s.out, tdTorList...)
			}
		}
	}
	// Stop the program only if all goroutines returned an error
	if arcSearchErr != nil && tdSearchErr != nil {
		fmt.Println("All searches returned an error.")
		log.WithFields(log.Fields{
			"input": s.in,
			"error": err,
		}).Fatal("All searches broke")
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

	// Final torrent we're working on as of now
	ft = s.out[index]

	// Download torrent and optionnaly open in torrent client
	switch ft.source {
	case "arc":
		getAndShowTorrent()
		openMagOrTorInClient(ft.filePath)
	case "td":
		ft.fileURL, ft.magnet, err = td.ExtractTorAndMag(ft.descURL)
		if err != nil {
			fmt.Println("An error occured while retrieving magnet and torrent file.")
			log.WithFields(log.Fields{
				"descURL":         ft.descURL,
				"sourcesToLookup": s.sourcesToLookup,
				"error":           err,
			}).Fatal("Could not retrieve magnet and torrent file")
		}
		switch {
		case ft.fileURL == "" && ft.magnet != "":
			log.WithFields(log.Fields{
				"torrentURL": ft.fileURL,
			}).Debug("Could not find a torrent file but successfully fetched a magnet link on the description page")
			getAndShowMagnet()
		case ft.fileURL != "" && ft.magnet == "":
			log.WithFields(log.Fields{
				"magnetLink": ft.magnet,
			}).Debug("Could not find a magnet link but successfully fetched a torrent file on the description page")
			getAndShowTorrent()
		default:
			log.WithFields(log.Fields{
				"torrentURL": ft.fileURL,
				"magnetLink": ft.magnet,
			}).Debug("Successfully fetched a torrent file and a magnet link on the description page")
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
	}

}
