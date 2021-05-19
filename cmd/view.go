package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
	"strings"
	"time"
)

var dbPath string

func init() {
	viewCmd.Flags().StringVarP(&dbPath, "path", "p", ".", "path to dir where db located")
	rootCmd.AddCommand(viewCmd)
}

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Views the server database",
	Long:  `Views the server database, the server must not be running when this is running`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := bolt.Open(strings.TrimSuffix(dbPath, "/")+"/spotify.db", 0666, &bolt.Options{Timeout: 1 * time.Second})
		if err != nil {
			return err
		}
		defer db.Close()

		err = db.View(func(tx *bolt.Tx) error {
			c := tx.Cursor()
			dumpCursor(tx, c, 0)
			return nil
		})
		if err != nil {
			return err
		}

		return nil
	},
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
