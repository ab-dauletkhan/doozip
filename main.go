package main

import (
	"log"

	"doozip/cmd/doozip"
)

func main() {
	if err := doozip.Run(); err != nil {
		log.Fatal(err)
	}
}
