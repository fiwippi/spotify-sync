package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func dumpCursor(tx *bolt.Tx, c *bolt.Cursor, indent int) {
	for k, v := c.First(); k != nil; k, v = c.Next() {
		if v == nil {
			fmt.Printf(strings.Repeat("\t", indent)+"[%s]\n", k)
			newBucket := c.Bucket().Bucket(k)
			if newBucket == nil {
				newBucket = tx.Bucket(k)
			}
			newCursor := newBucket.Cursor()
			dumpCursor(tx, newCursor, indent+1)
		} else {
			fmt.Printf(strings.Repeat("\t", indent)+"%s\n", k)
			fmt.Printf(strings.Repeat("\t", indent+1)+"%s\n", v)
		}
	}
}

func main() {
	db, err := bolt.Open("spotify.db", 0666, &bolt.Options{Timeout: 1 * time.Second})
	check(err)
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		c := tx.Cursor()
		dumpCursor(tx, c, 0)
		return nil
	})
	check(err)
}