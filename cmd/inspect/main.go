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
	limit := flag.Int("limit", -1, "")
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

	inspect(db, title, *limit)
}

func inspect(db *bolt.DB, title string, limit int) {
	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(title))

		if bucket == nil {
			fmt.Printf("No entry for title: '%s'\n", title)
			return nil
		}

		fmt.Printf("title: '%s'\n", title)
		dump(bucket, limit, "")

		return nil
	})
}

func dump(bucket *bolt.Bucket, limit int, prefix string) {
	bucket.ForEach(func(key, val []byte) error {
		if val == nil {
			fmt.Printf("%sbucket: %#v\n", prefix, string(key))
			dump(bucket.Bucket(key), limit, prefix+"\t")
		} else {
			stringVal := string(val)
			if limit > 0 && len(stringVal) > limit {
				stringVal = stringVal[:limit]
			}

			fmt.Printf("%skey: %#v, value: %#v\n", prefix, string(key), stringVal)
		}

		return nil
	})
}
