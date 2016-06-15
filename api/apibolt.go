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

	// check if the title redirects
	title, err := bl.lookupRedirect(title)
	if err != nil {
		return Page{}, err
	}

	links, err := bl.lookupLinks(title)
	if err != nil {
		return Page{}, err
	}

	return Page{redirector, title, links}, nil
}

// Checks if the given title redirects to a different page.
// If it does, returns the title that is redirected to.
// If it doesn't, returns the original title.
func (bl *boltLoader) lookupRedirect(title string) (string, error) {
	titleBytes := []byte(title)

	err := bl.redir.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(titleBytes)

		// no bucket means no redirect
		if bucket == nil {
			return nil
		}

		// the redirect bucket contains only one key, value pair, the title to redirect to
		bucket.ForEach(func(key, value []byte) error {
			titleBytes = []byte(EncodeTitle(string(key)))
			return nil
		})
		return nil
	})

	return string(titleBytes), err
}

func (bl *boltLoader) lookupLinks(title string) ([]string, error) {
	titleBytes := []byte(title)

	// load up the links
	var bytesLinks [][]byte
	err := bl.index.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(titleBytes)

		// no links for this page?
		if bucket == nil {
			return errors.New("No entry for title '" + string(titleBytes) + "'")
		}

		// else, each key is a link, so grab them all
		bucket.ForEach(func(key, value []byte) error {
			bytesLinks = append(bytesLinks, key)
			return nil
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	// convert the links to strings
	var links []string
	for _, bytesLink := range bytesLinks {
		links = append(links, string(bytesLink))
	}

	return links, nil
}
