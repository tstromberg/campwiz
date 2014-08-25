// The "main" package contains the command-line utility functions.
package main

import (
	"autocamper"
)

func main() {
	resp, err := campquery.Search("94122")
	if err != nil {
		log.Fatalf("Fetch error: %s", err)
	}
	err = campquery.Parse(resp.Body)
	if err != nil {
		log.Fatalf("Parse error: %s", err)
	}
}
