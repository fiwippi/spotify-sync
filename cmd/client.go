package main

import (
	"log"
	"spotify-sync/pkg/client"
)

func main() {
	if err := client.NewClient().Run(); err != nil {
		log.Fatal(err)
	}
}
