package wiki

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/boltdb/bolt"
)

const DefaultIndexName = "db/index.db"
const defaultFileMode = 0600

var redirectKey = []byte("redir")
var linksKey = []byte("links")

const linkSeparator = "\n"

type boltLoader struct {
	// connection to db of {title -> links} mappings
	index *bolt.DB

	// waitgroup to keep track of whether the connections are in use
	wg sync.WaitGroup

	// atomic boolean to block new loads from starting when a close is requested
	closing   bool
	closeLock sync.Mutex
}

func GetBoltPageLoader() (PageLoader, error) {
	return getBoltConnection(DefaultIndexName, defaultFileMode, &bolt.Options{ReadOnly: true})
}

func GetBoltPageSaver(indexFilename string) (PageSaver, error) {
	return getBoltConnection(indexFilename, defaultFileMode, nil)
}

func GetBoltPageRepository(indexFilename string) (PageRepository, error) {
	return getBoltConnection(indexFilename, defaultFileMode, nil)
}

func getBoltConnection(indexFilename string, mode os.FileMode, options *bolt.Options) (*boltLoader, error) {
	index, err := bolt.Open(indexFilename, mode, options)
	if err != nil {
		return nil, err
	}

	pageLoader := boltLoader{index: index}
	return &pageLoader, nil
}

func (bl *boltLoader) LoadPage(title string) (Page, error) {
	// make sure the connections don't close until we're done
	bl.wg.Add(1)
	defer bl.wg.Done()

	if bl.isClosing() {
		return Page{}, errors.New("Connection closed")
	}

	var page Page

	err := bl.index.View(func(tx *bolt.Tx) error {
		var viewErr error
		page, viewErr = bl.loadPageWithRedirect(tx, title)
		return viewErr
	})

	if err != nil {
		return Page{}, err
	}

	return page, nil
}

func (bl *boltLoader) LoadPages(titles []string) ([]Page, error) {
	var pages []Page

	for _, title := range titles {
		page, err := bl.LoadPage(title)
		if err != nil {
			return nil, err
		}

		pages = append(pages, page)
	}

	return pages, nil
}

func (bl *boltLoader) loadPageWithRedirect(tx *bolt.Tx, title string) (Page, error) {
	page, err := bl.loadPage(tx, title)
	if err != nil {
		return Page{}, err
	}

	// check if the title redirects
	if page.Redirect != "" {
		page, err = bl.loadPage(tx, page.Redirect)
		if err != nil {
			return Page{}, err
		}
		page.Redirector = title
	}

	return page, nil
}

func (bl *boltLoader) loadPage(tx *bolt.Tx, title string) (Page, error) {
	bucket := tx.Bucket([]byte(title))

	if bucket == nil {
		return Page{}, errors.New("No entry for title '" + title + "'")
	}

	redirect := string(bucket.Get(redirectKey))
	links := decodeLinks(bucket.Get(linksKey))

	return Page{Title: title, Redirect: redirect, Links: links}, nil
}

func (bl *boltLoader) SavePage(page Page) error {
	err := bl.index.Update(func(tx *bolt.Tx) error {
		return bl.savePage(tx, page)
	})

	return err
}

func (bl *boltLoader) SavePages(pages []Page) error {
	err := bl.index.Update(func(tx *bolt.Tx) error {
		for _, page := range pages {
			err := bl.savePage(tx, page)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

func (bl *boltLoader) savePage(tx *bolt.Tx, page Page) error {
	bucket, err := tx.CreateBucketIfNotExists([]byte(page.Title))
	if err != nil {
		return fmt.Errorf("error while creating bucket for title '%s': '%v'", page.Title, err)
	}

	if page.Redirect != "" {
		err = bucket.Put(redirectKey, []byte(page.Redirect))
		if err != nil {
			return err
		}
	}

	if len(page.Links) != 0 {
		err = bucket.Put(linksKey, encodeLinks(page.Links))
		if err != nil {
			return err
		}
	}

	return nil
}

func (bl *boltLoader) FirstPage() (Page, error) {
	return Page{}, errors.New("not implemented")
}

func (bl *boltLoader) NextPage(title string) (Page, error) {
	return Page{}, errors.New("not implemented")
}

func (bl *boltLoader) NextPages(title string, count int) ([]Page, error) {
	return nil, errors.New("not implemented")
}

// Blocks new loads from starting, waits for existing loads to complete,
// and then shuts down the db connections
func (bl *boltLoader) Close() error {
	// set the closing flag so that no new loads are started
	bl.setClosing()

	// wait until existing loads are done
	bl.wg.Wait()

	// then shut down the connections
	bl.index.Close()

	return nil
}

func (bl *boltLoader) setClosing() {
	bl.closeLock.Lock()
	defer bl.closeLock.Unlock()

	bl.closing = true
}

func (bl *boltLoader) isClosing() bool {
	bl.closeLock.Lock()
	defer bl.closeLock.Unlock()

	return bl.closing
}

func encodeLinks(links []string) []byte {
	return []byte(strings.Join(links, linkSeparator))
}

func decodeLinks(encodedLinks []byte) []string {
	return strings.Split(string(encodedLinks), linkSeparator)
}
