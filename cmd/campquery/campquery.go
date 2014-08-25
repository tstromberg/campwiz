// The "main" package contains the command-line utility functions.
package main

import (
	"log"

	"github.com/tstromberg/autocamper"
)

func main() {
	resp, err := autocamper.Search("94122")
	if err != nil {
		log.Fatalf("Fetch error: %s", err)
	}
	err = autocamper.Parse(resp.Body)
	if err != nil {
		log.Fatalf("Parse error: %s", err)
	}
}
