package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/boltdb/bolt"
	"github.com/kbuzsaki/wikidegree/wiki"
)

func main() {
	dbFilename := flag.String("db", wiki.DefaultIndexName, "the boltdb file")
	bareTitle := flag.String("title", "", "")
	bare := flag.Bool("bare", false, "")
	flag.Parse()

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	db, err := bolt.Open(*dbFilename, 0600, &bolt.Options{ReadOnly: true})
	if err != nil {
		log.Fatal(err)
	}

	var title string
	if *bare {
		title = *bareTitle
	} else {
		title = wiki.NormalizeTitle(*bareTitle)
	}

	inspect(db, title)
}

func inspect(db *bolt.DB, title string) {
	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(title))

		if bucket == nil {
			fmt.Printf("No entry for title: '%s'\n", title)
			return nil
		}

		fmt.Printf("title: '%s'\n", title)
		bucket.ForEach(func(key, value []byte) error {
			fmt.Printf("key: %#v, value: %#v\n", string(key), string(value))
			return nil
		})

		return nil
	})
}
