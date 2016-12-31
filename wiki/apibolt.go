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
var linkersKey = []byte("linkers")
var blobBucketKey = []byte("blob")

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
	if err := bl.retain(); err != nil {
		return Page{}, err
	}
	defer bl.release()

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
	if err := bl.retain(); err != nil {
		return nil, err
	}
	defer bl.release()

	var pages []Page
	err := bl.index.View(func(tx *bolt.Tx) error {
		for _, title := range titles {
			page, err := bl.loadPageWithRedirect(tx, title)
			if err != nil {
				return err
			}
			pages = append(pages, page)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return pages, nil
}

func (bl *boltLoader) loadPageWithRedirect(tx *bolt.Tx, title string) (Page, error) {
	page, err := bl.loadPage(tx, title)
	if err != nil {
		return Page{}, fmt.Errorf("error loading page %#v: %s", title, err)
	}

	// check if the title redirects
	if page.Redirect != "" {
		redirectTitle := page.Redirect
		page, err = bl.loadPage(tx, redirectTitle)
		if err != nil {
			return Page{}, fmt.Errorf("error loading redirect from %#v to %#v: %s", title, redirectTitle, err)
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

	redirect := decodeTitle(bucket.Get(redirectKey))
	links := decodeLinks(bucket.Get(linksKey))
	linkers := decodeLinks(bucket.Get(linkersKey))
	blob, err := loadBlob(bucket, blobBucketKey)
	if err != nil {
		return Page{}, fmt.Errorf("Unable to load blob for page '%s', error: %s", title, err)
	}

	return Page{Title: title, Redirect: redirect, Links: links, Linkers: linkers, Blob: blob}, nil
}

func loadBlob(bucket *bolt.Bucket, key []byte) (Blob, error) {
	blobBucket := bucket.Bucket(key)
	if blobBucket == nil {
		return nil, nil
	}

	blob := make(Blob)
	err := blobBucket.ForEach(func (key, val []byte) error {
		blob[string(key)] = append([]byte(nil), val...)
		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(blob) == 0 {
		return nil, nil
	}

	return blob, nil
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

	err = putOrDelete(bucket, redirectKey, encodeTitle(page.Redirect))
	if err != nil {
		return err
	}

	err = putOrDelete(bucket, linksKey, encodeLinks(page.Links))
	if err != nil {
		return err
	}

	err = putOrDelete(bucket, linkersKey, encodeLinks(page.Linkers))
	if err != nil {
		return err
	}

	if len(page.Blob) != 0 {
		blobBucket, err := bucket.CreateBucketIfNotExists(blobBucketKey)
		if err != nil {
			return err
		}

		putOrDeleteBlobEntries(blobBucket, page.Blob)
	}

	return nil
}

func (bl *boltLoader) SavePageBlobs(pages []Page) error {
	err := bl.index.Update(func(tx *bolt.Tx) error {
		for _, page := range pages {
			err := bl.savePageBlob(tx, page)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

func (bl *boltLoader) savePageBlob(tx *bolt.Tx, page Page) error {
	bucket, err := tx.CreateBucketIfNotExists([]byte(page.Title))
	if err != nil {
		return fmt.Errorf("error while creating bucket for title '%s': '%v'", page.Title, err)
	}

	if len(page.Blob) != 0 {
		blobBucket, err := bucket.CreateBucketIfNotExists(blobBucketKey)
		if err != nil {
			return err
		}

		putOrDeleteBlobEntries(blobBucket, page.Blob)
	}

	return nil
}

func putOrDelete(bucket *bolt.Bucket, key []byte, val []byte) error {
	if val == nil {
		err := bucket.Delete(key)
		if err != nil {
			return err
		}
	} else {
		err := bucket.Put(key, val)
		if err != nil {
			return err
		}
	}

	return nil
}

func putOrDeleteBlobEntries(blobBucket *bolt.Bucket, blob Blob) error {
	for key, val := range blob {
		err := putOrDelete(blobBucket, []byte(key), val)
		if err != nil {
			return err
		}
	}

	return nil
}

func (bl *boltLoader) FirstPage() (Page, error) {
	return bl.NextPage("")
}

func (bl *boltLoader) NextPage(title string) (Page, error) {
	pages, err := bl.NextPages(title, 1)
	if err != nil {
		return Page{}, err
	}

	if len(pages) < 1 {
		return Page{}, nil
	}

	return pages[0], nil
}

func (bl *boltLoader) NextPages(title string, count int) ([]Page, error) {
	if count < 1 {
		return nil, fmt.Errorf("count must be positive, was %d", count)
	}

	if err := bl.retain(); err != nil {
		return nil, err
	}
	defer bl.release()

	var pages []Page
	err := bl.index.View(func(tx *bolt.Tx) error {
		cursor := tx.Cursor()

		//
		key, val := cursor.Seek([]byte(title))
		if string(key) == title {
			key, val = cursor.Next()
		}

		for ; key != nil; key, val = cursor.Next() {
			// a nil value means that the key is to a bucket, which is a page
			if val == nil {
				page, err := bl.loadPage(tx, string(key))
				if err != nil {
					return err
				}

				pages = append(pages, page)
				if len(pages) >= count {
					return nil
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return pages, nil
}

func (bl *boltLoader) NextTitles(title string, count int) ([]string, error) {
	if count < 1 {
		return nil, fmt.Errorf("count must be positive, was %d", count)
	}

	if err := bl.retain(); err != nil {
		return nil, err
	}
	defer bl.release()

	var titles []string
	err := bl.index.View(func(tx *bolt.Tx) error {
		cursor := tx.Cursor()

		key, val := cursor.Seek([]byte(title))
		if string(key) == title {
			key, val = cursor.Next()
		}

		for ; key != nil; key, val = cursor.Next() {
			// a nil value means that the key is to a bucket, which is a page
			if val == nil {
				titles = append(titles, string(key))
				if len(titles) >= count {
					return nil
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return titles, nil
}

func (bl *boltLoader) DeleteTitle(title string) error {
	if err := bl.retain(); err != nil {
		return err
	}
	defer bl.release()

	err := bl.index.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(title))
	})

	return err
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

func (bl *boltLoader) retain() error {
	bl.wg.Add(1)

	if bl.isClosing() {
		bl.wg.Done()
		return errors.New("Connection closed")
	}

	return nil
}

func (bl *boltLoader) release() {
	bl.wg.Done()
}

// TODO: use these helpers for the actual page title, not just redirects
func encodeTitle(title string) []byte {
	if title == "" {
		return nil
	}

	return []byte(title)
}

func decodeTitle(encodedTitle []byte) string {
	if len(encodedTitle) == 0 {
		return ""
	}

	return string(encodedTitle)
}

func encodeLinks(links []string) []byte {
	if links == nil {
		return nil
	} else if len(links) == 0 {
		return []byte{}
	}
	return []byte(strings.Join(links, linkSeparator))
}

func decodeLinks(encodedLinks []byte) []string {
	if encodedLinks == nil {
		return nil
	} else if len(encodedLinks) == 0 {
		return []string{}
	}
	return strings.Split(string(encodedLinks), linkSeparator)
}
