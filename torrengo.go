package main

import (
	"log"

	"github.com/juliensalinas/torrengo/arc"
)

// func checkInput(in string) error {

// 	if in == "" {
// 		return fmt.Errorf("user input should not be empty")
// 	}

// 	return nil
// }

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	torrents, err := arc.Search("Dumas")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%v", torrents)

}
