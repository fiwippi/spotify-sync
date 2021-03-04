package main

import (
	"github.com/fiwippi/spotify-sync/pkg/client"
	"log"
)

func main() {
	if err := client.NewClient().Run(); err != nil {
		log.Fatal(err)
	}
}
