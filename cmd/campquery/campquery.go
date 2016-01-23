// The "main" package contains the command-line utility functions.
package main

import (
	"log"

	"github.com/tstromberg/autocamper/query"
)

func main() {
	resp, err := query.Search("94122")
	if err != nil {
		log.Fatalf("Fetch error: %s", err)
	}
	err = query.Parse(resp.Body)
	if err != nil {
		log.Fatalf("Parse error: %s", err)
	}
}
