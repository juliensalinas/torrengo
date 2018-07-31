package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/onrik/logrus/filename"
	log "github.com/sirupsen/logrus"

	"github.com/juliensalinas/torrengo/arc"
	"github.com/juliensalinas/torrengo/otts"
	"github.com/juliensalinas/torrengo/td"
	"github.com/juliensalinas/torrengo/tpb"
	"github.com/olekukonko/tablewriter"
)

// lineBreak sets the OS dependent line break (initialized in init())
var lineBreak string

// sources maps source short names to real names
var sources = map[string]string{
	"arc":  "Archive",
	"td":   "Torrent Downloads",
	"tpb":  "The Pirate Bay",
	"otts": "1337x",
}

// ft is the final torrent the user wants to download
var ft torrent

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
			sources[t.source],
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

// getTorrentFile retrieves and displays torrent file to user
func getTorrentFile() {
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
}

func openMagOrTorInClient(resource string) {
	// Open torrent in client
	log.WithFields(log.Fields{
		"resource": resource,
		"client":   "Deluge",
	}).Debug("Opening magnet link or torrent file with torrent client")
	fmt.Println("Opening torrent in client...")
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
	log.SetLevel(log.ErrorLevel)

	// Log filename and line number.
	// Should be removed from production because adds a performance cost.
	log.AddHook(filename.NewHook())

	// Set custom line break in order for the script to work on any OS
	if runtime.GOOS == "windows" {
		lineBreak = "\r\n"
	} else {
		lineBreak = "\n"
	}
}

// TODO: tpb is changing very frequently so implement a proxy lookup
// TODO: set a timeout per search goroutine
func main() {

	// Get command line flags and arguments
	usrSourcesPtr := flag.String("w", "all", "A comma separated list of websites "+
		"you want to search (e.g. arc,td,tbp). Choices: arc | td | tpb | all. "+
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
	usrSourcesSlc := strings.Split(*usrSourcesPtr, ",")
	cleanedUsrSourcesSlc := rmDuplicates(usrSourcesSlc)
	for _, usrSource := range cleanedUsrSourcesSlc {
		if usrSource == "all" {
			cleanedUsrSourcesSlc = []string{"arc", "td", "tpb", "otts"}
			break
		}
		if usrSource != "arc" && usrSource != "td" && usrSource != "tpb" && usrSource != "otts" {
			fmt.Printf("This website is not correct: %v%v", usrSource, lineBreak)
			log.WithFields(log.Fields{
				"sourcesList": cleanedUsrSourcesSlc,
				"wrongSource": usrSource,
			}).Fatal("Unknown source in user sources list")
		}
	}
	s := search{
		in:              strings.Join(args, " "),
		sourcesToLookup: cleanedUsrSourcesSlc,
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
	tpbTorListCh := make(chan []torrent)
	ottsTorListCh := make(chan []torrent)

	// Channels for errors
	arcSearchErrCh := make(chan error)
	tdSearchErrCh := make(chan error)
	tpbSearchErrCh := make(chan error)
	ottsSearchErrCh := make(chan error)

	// Launch all torrent search goroutines
	for _, source := range s.sourcesToLookup {
		switch source {
		// User wants to search arc
		case "arc":
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
						size:     "Unknown",
						leechers: -1,
						seeders:  -1,
						source:   "arc",
					}
					torList = append(torList, t)
				}
				arcTorListCh <- torList
			}()

		// User wants to search td
		case "td":
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

		// User wants to search tpb
		case "tpb":
			go func() {
				tpbTorrents, err := tpb.Lookup(s.in)
				if err != nil {
					tpbSearchErrCh <- err
				}
				var torList []torrent
				for _, tpbTorrent := range tpbTorrents {
					t := torrent{
						magnet:   tpbTorrent.Magnet,
						name:     tpbTorrent.Name,
						size:     tpbTorrent.Size,
						uplDate:  tpbTorrent.UplDate,
						leechers: tpbTorrent.Leechers,
						seeders:  tpbTorrent.Seeders,
						source:   "tpb",
					}
					torList = append(torList, t)
				}
				tpbTorListCh <- torList
			}()
		// User wants to search otts
		case "otts":
			go func() {
				ottsTorrents, err := otts.Lookup(s.in)
				if err != nil {
					ottsSearchErrCh <- err
				}
				var torList []torrent
				for _, ottsTorrent := range ottsTorrents {
					t := torrent{
						descURL:  ottsTorrent.DescURL,
						name:     ottsTorrent.Name,
						size:     ottsTorrent.Size,
						uplDate:  ottsTorrent.UplDate,
						leechers: ottsTorrent.Leechers,
						seeders:  ottsTorrent.Seeders,
						source:   "otts",
					}
					torList = append(torList, t)
				}
				ottsTorListCh <- torList
			}()
		}
	}

	// Initialize search errors
	var tdSearchErr, arcSearchErr, tpbSearchErr, ottsSearchErr error

	// Gather all goroutines results
	for _, source := range s.sourcesToLookup {
		switch source {
		case "arc":
			// Get results or error from arc
			select {
			case arcSearchErr = <-arcSearchErrCh:
				fmt.Printf("An error occured during search on %v%v", sources["arc"], lineBreak)
				log.WithFields(log.Fields{
					"input": s.in,
					"error": err,
				}).Error("The arc search goroutine broke")
			case arcTorList := <-arcTorListCh:
				s.out = append(s.out, arcTorList...)
			}
		case "td":
			// Get results or error from td
			select {
			case tdSearchErr = <-tdSearchErrCh:
				fmt.Printf("An error occured during search on %v%v", sources["td"], lineBreak)
				log.WithFields(log.Fields{
					"input": s.in,
					"error": err,
				}).Error("The td search goroutine broke")
			case tdTorList := <-tdTorListCh:
				s.out = append(s.out, tdTorList...)
			}
		case "tpb":
			// Get results or error from tpb
			select {
			case tpbSearchErr = <-tpbSearchErrCh:
				fmt.Printf("An error occured during search on %v%v", sources["tpb"], lineBreak)
				log.WithFields(log.Fields{
					"input": s.in,
					"error": err,
				}).Error("The tpb search goroutine broke")
			case tpbTorList := <-tpbTorListCh:
				s.out = append(s.out, tpbTorList...)
			}
		case "otts":
			// Get results or error from otts
			select {
			case ottsSearchErr = <-ottsSearchErrCh:
				fmt.Printf("An error occured during search on %v%v", sources["otts"], lineBreak)
				log.WithFields(log.Fields{
					"input": s.in,
					"error": err,
				}).Error("The otts search goroutine broke")
			case ottsTorList := <-ottsTorListCh:
				s.out = append(s.out, ottsTorList...)
			}
		}
	}
	// Stop the program only if all goroutines returned an error
	if arcSearchErr != nil && tdSearchErr != nil && tpbSearchErr != nil && ottsSearchErr != nil {
		fmt.Println("All searches returned an error.")
		log.WithFields(log.Fields{
			"input": s.in,
			"error": err,
		}).Fatal("All searches broke")
	}

	// Stop the program if no result found
	if len(s.out) == 0 {
		fmt.Println("No result found...")
		os.Exit(1)
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
		indexStr, err := reader.ReadString('\n') // returns string + delimiter
		if err != nil {
			fmt.Println("Could not read your input, please try again (should be an integer):")
			continue
		}
		// Remove delimiter which depends on OS + white spaces if any, and convert to integer
		index, err = strconv.Atoi(strings.TrimSpace(strings.TrimSuffix(indexStr, lineBreak)))
		if err != nil {
			fmt.Println("Please enter an integer:")
			continue
		}
		break
	}

	// Final torrent we're working on as of now
	ft = s.out[index]

	// Read from user input whether he wants to open torrent in client or not
	reader = bufio.NewReader(os.Stdin)
	fmt.Println("Do you want to open torrent in Deluge client? y / n")
	var launchClient string
	for {
		launchClientStr, err := reader.ReadString('\n') // returns string + delimiter
		if err != nil {
			fmt.Println("Could not read your input, please try again (should be 'y' or 'n'):")
			continue
		}
		// Remove delimiter which depends on OS + white spaces if any
		launchClient = strings.TrimSpace(strings.TrimSuffix(launchClientStr, lineBreak))
		break
	}

	// Download torrent and optionnaly open in torrent client
	switch ft.source {
	case "arc":
		getTorrentFile()
		fmt.Printf("Here is your torrent file: %s%s%s", lineBreak, ft.filePath, lineBreak)
		if launchClient == "y" {
			openMagOrTorInClient(ft.filePath)
		}
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
			fmt.Printf("Here is your magnet link: %s%s%s", lineBreak, ft.magnet, lineBreak)
			if launchClient == "y" {
				openMagOrTorInClient(ft.magnet)
			}
		case ft.fileURL != "" && ft.magnet == "":
			log.WithFields(log.Fields{
				"magnetLink": ft.magnet,
			}).Debug("Could not find a magnet link but successfully fetched a torrent file on the description page")
			getTorrentFile()
			fmt.Printf("Here is your torrent file: %s%s%s", lineBreak, ft.filePath, lineBreak)
			if launchClient == "y" {
				openMagOrTorInClient(ft.filePath)
			}
		default:
			log.WithFields(log.Fields{
				"torrentURL": ft.fileURL,
				"magnetLink": ft.magnet,
			}).Debug("Successfully fetched a torrent file and a magnet link on the description page")
			// Ask user to choose between file download and magnet download
			reader := bufio.NewReader(os.Stdin)
			fmt.Println("We found a torrent file and a magnet link, which one would you like to download?" +
				lineBreak + "1) Magnet link" + lineBreak + "2) Torrent file (careful: not working 100% of the time)")
			var choice int
			for {
				choiceStr, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println("Could not read your input, please enter your choice (1 or 2):")
					continue
				}
				choice, err = strconv.Atoi(strings.TrimSpace(strings.TrimSuffix(choiceStr, lineBreak)))
				if err != nil {
					fmt.Println("Please enter an integer:")
					continue
				}
				break
			}
			switch choice {
			case 1:
				fmt.Printf("Here is your magnet link: %s%s%s", lineBreak, ft.magnet, lineBreak)
				if launchClient == "y" {
					openMagOrTorInClient(ft.magnet)
				}
			case 2:
				getTorrentFile()
				fmt.Printf("Here is your torrent file: %s%s%s", lineBreak, ft.filePath, lineBreak)
				if launchClient == "y" {
					openMagOrTorInClient(ft.filePath)
				}
			}
		}
	case "tpb":
		fmt.Printf("Here is your magnet link: %s%s%s", lineBreak, ft.magnet, lineBreak)
		if launchClient == "y" {
			openMagOrTorInClient(ft.magnet)
		}
	case "otts":
		ft.magnet, err = otts.ExtractMag(ft.descURL)
		if err != nil {
			fmt.Println("An error occured while retrieving magnet.")
			log.WithFields(log.Fields{
				"descURL":         ft.descURL,
				"sourcesToLookup": s.sourcesToLookup,
				"error":           err,
			}).Fatal("Could not retrieve magnet")
		}
		fmt.Printf("Here is your magnet link: %s%s%s", lineBreak, ft.magnet, lineBreak)
		if launchClient == "y" {
			openMagOrTorInClient(ft.magnet)
		}
	}
}
