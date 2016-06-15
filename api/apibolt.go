package api

import (
	"errors"
	"github.com/boltdb/bolt"
	"sync"
)

const defaultIndexName = "db/index.db"
const defaultRedirName = "db/redir.db"

type boltLoader struct {
	// connection to db of {title -> links} mappings
	index *bolt.DB
	// connection to db of {title -> redirect} mappings
	redir *bolt.DB

	// waitgroup to keep track of whether the connections are in use
	wg sync.WaitGroup

	// atomic boolean to block new loads from starting when a close is requested
	closing bool
	closeLock sync.Mutex
}

func GetBoltPageLoader() (PageLoader, error) {
	index, err := bolt.Open(defaultIndexName, 0600, &bolt.Options{ReadOnly: true})
	if err != nil {
		return nil, err
	}

	redir, err := bolt.Open(defaultRedirName, 0600, &bolt.Options{ReadOnly: true})
	if err != nil {
		return nil, err
	}

	pageLoader := boltLoader{index, redir, sync.WaitGroup{}, false, sync.Mutex{}}
	return &pageLoader, nil
}

// Blocks new loads from starting, waits for existing loads to complete,
// and then shuts down the db connections
func (bl *boltLoader) Close() {
	// set the closing flag so that no new loads are started
	bl.setClosing()

	// wait until existing loads are done
	bl.wg.Wait()

	// then shut down the connections
	bl.index.Close()
	bl.redir.Close()
}

func (bl *boltLoader) setClosing() {
	bl.closeLock.Lock()
	defer bl.closeLock.Unlock()

	bl.closing = true
}

func (bl *boltLoader) IsClosing() bool {
	bl.closeLock.Lock()
	defer bl.closeLock.Unlock()

	return bl.closing
}

func (bl *boltLoader) LoadPage(title string) (Page, error) {
	// make sure the connections don't close until we're done
	bl.wg.Add(1)
	defer bl.wg.Done()

	if bl.IsClosing() {
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
	var links []string
	err := bl.index.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(titleBytes)

		// no links for this page?
		if bucket == nil {
			return errors.New("No entry for title '" + string(titleBytes) + "'")
		}

		// else, each key is a link, so grab them all
		bucket.ForEach(func(key, value []byte) error {
			links = append(links, string(key))
			return nil
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return links, nil
}
