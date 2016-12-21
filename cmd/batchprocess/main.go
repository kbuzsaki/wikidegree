package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"strings"

	"sync"

	"github.com/kbuzsaki/wikidegree/wiki"
)

const printThresh = 1000000
const defaultBatchSize = 1000
const defaultConcurrency = 1

var debug bool

func main() {
	flag.BoolVar(&debug, "debug", false, "print debug output")
	dbFilename := flag.String("db", wiki.DefaultIndexName, "the boltdb file")
	batchSize := flag.Int("batch", defaultBatchSize, "number of pages to pass to the processing function at a time")
	concurrency := flag.Int("concurrency", defaultConcurrency, "number of goroutines to use for processing")
	flag.Parse()

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	pr, err := wiki.GetBoltPageRepository(*dbFilename)
	if err != nil {
		log.Fatal(err)
	}

	err = process(pr, *batchSize, *concurrency, cleanDeadLinks)
	if err != nil {
		log.Fatal(err)
	}
}

type processor func(pr wiki.PageRepository, pages []wiki.Page) error

func process(pr wiki.PageRepository, batchSize int, concurrency int, f processor) error {
	pageBuffer, err := pr.NextPages("", batchSize)
	if err != nil {
		return err
	}

	wg := &sync.WaitGroup{}
	pageBuffers := make(chan []wiki.Page, 2*concurrency)
	errs := make(chan error)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go processWorker(pr, f, wg, pageBuffers, errs)
	}

	counter := 0
	for len(pageBuffer) != 0 {
		if debug && counter%printThresh == 0 {
			log.Println("processed", counter)
		}

		select {
		case pageBuffers <- pageBuffer:
			// pass
		case err := <-errs:
			close(pageBuffers)
			close(errs)
			return err
		}

		counter += len(pageBuffer)
		pageBuffer, err = pr.NextPages(pageBuffer[len(pageBuffer)-1].Title, batchSize)
	}

	wg.Wait()
	return nil
}

func processWorker(pr wiki.PageRepository, f processor, wg *sync.WaitGroup, pageBuffers <-chan []wiki.Page, errs chan<- error) {
	defer wg.Done()
	for pageBuffer := range pageBuffers {
		err := f(pr, pageBuffer)
		if err != nil {
			errs <- err
		}
	}
}

func printShortNames(pr wiki.PageRepository, pages []wiki.Page) error {
	for _, page := range pages {
		if len(page.Title) < 10 {
			fmt.Println(page.Title)
		}
	}

	return nil
}

func cleanDeadLinks(pr wiki.PageRepository, pages []wiki.Page) error {
	var updatedPages []wiki.Page

	for _, page := range pages {
		validLinks, updated, err := getValidLinks(pr, page)
		if err != nil {
			return err
		}

		if updated {
			page.Links = validLinks
			updatedPages = append(updatedPages, page)
		}
	}

	return pr.SavePages(updatedPages)
}

func getValidLinks(pr wiki.PageRepository, page wiki.Page) ([]string, bool, error) {
	var validLinks []string

	updated := false
	for _, link := range page.Links {
		_, err := pr.LoadPage(link)
		if err != nil && !strings.HasPrefix(err.Error(), "No entry") {
			return nil, false, err
		} else if err != nil {
			if debug {
				log.Printf("found dead link %#v in page %#v\n", link, page.Title)
			}
			updated = true
		} else {
			validLinks = append(validLinks, link)
		}
	}

	return validLinks, updated, nil
}
