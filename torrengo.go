package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/juliensalinas/torrengo/arc"
)

func clean(in string) (string, error) {

	// Clean user input by removing useless spaces.
	clIn := strings.TrimSpace(in)

	// If user input is empty raise an error.
	if clIn == "" {
		return "", fmt.Errorf("user input should not be empty")
	}

	return clIn, nil
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

	var torrents [][2]string

	switch *websitePtr {
	case "archive":
		torrents, err = arc.Search(clIn)
		if err != nil {
			log.Fatal(err)
		}
	case "all":
		fmt.Println("all")
	}

	log.Printf("%v", torrents)

}
