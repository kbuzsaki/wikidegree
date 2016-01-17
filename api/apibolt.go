package api

import (
	"errors"
	"github.com/boltdb/bolt"
	"log"
	"sync"
)

const defaultIndexName = "db/index.db"
const defaultRedirName = "db/redir.db"

type boltLoader struct {
	index *bolt.DB
	redir *bolt.DB
	wg sync.WaitGroup
	closing bool
}

func GetBoltPageLoader() (PageLoader, error) {
	index, err := bolt.Open(defaultIndexName, 0600, nil)
	if err != nil {
		return nil, err
	}

	redir, err := bolt.Open(defaultRedirName, 0600, nil)
	if err != nil {
		return nil, err
	}

	pageLoader := boltLoader{index, redir, sync.WaitGroup{}, false}
	return &pageLoader, nil
}

func (bl *boltLoader) Close() {
	// wait until everyone is done before closing bolt
	// this is kinda hacky and likely not the right way to do things...
	bl.closing = true
	bl.wg.Wait()

	bl.index.Close()
	bl.redir.Close()
}

func (bl *boltLoader) LoadPage(title string) (Page, error) {
	// make sure the connections don't close until we're done
	bl.wg.Add(1)
	defer bl.wg.Done()

	if bl.closing {
		return Page{}, errors.New("Connection closed")
	}

	// preserve the original link even if there's a redirect
	redirector := title

	titleBytes := []byte(title)

	// check if the title redirects
	err := bl.redir.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(titleBytes)
		// no redirect for this page
        if bucket == nil {
			return nil
		}

		bucket.ForEach(func(key, value []byte) error {
			// if we find a redirect, switch to that instead
			log.Println("Redirecting", title, "to", string(key))
			titleBytes = []byte(EncodeTitle(string(key)))
			return nil
		})
		return nil
	})
	if err != nil {
		return Page{}, err
	}

	// load up the links
	var bytesLinks [][]byte
	err = bl.index.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(titleBytes)
		// no links for this page?
        if bucket == nil {
			return errors.New("No entry for title '" + string(titleBytes) + "'")
		}

		bucket.ForEach(func(key, value []byte) error {
			bytesLinks = append(bytesLinks, key)
			return nil
		})
		return nil
	})
	if err != nil {
		return Page{}, err
	}

	// convert the links to strings
	var links []string
	for _, bytesLink := range bytesLinks {
		links = append(links, string(bytesLink))
	}

	return Page{redirector, string(titleBytes), links}, nil
}
