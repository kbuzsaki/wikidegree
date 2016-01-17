package api

import (
	"errors"
	"github.com/boltdb/bolt"
)

const defaultIndexName = "db/index.db"
const defaultRedirName = "db/redir.db"

type boltLoader struct {
	index *bolt.DB
	redir *bolt.DB
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

	pageLoader := boltLoader{index, redir}
	return &pageLoader, nil
}

func (bl *boltLoader) Close() {
	bl.index.Close()
	bl.redir.Close()
}

func (bl *boltLoader) LoadPage(title string) (Page, error) {
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
			titleBytes = key
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

	return Page{title, links}, nil
}
