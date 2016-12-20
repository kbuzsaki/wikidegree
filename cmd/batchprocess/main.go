package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/kbuzsaki/wikidegree/wiki"
)

const defaultBatchSize = 1000

func main() {
	dbFilename := flag.String("db", wiki.DefaultIndexName, "the boltdb file")
	batchSize := flag.Int("batch", defaultBatchSize, "number of pages to pass to the processing function at a time")
	flag.Parse()

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	pr, err := wiki.GetBoltPageRepository(*dbFilename)
	if err != nil {
		log.Fatal(err)
	}

	err = process(pr, *batchSize, printShortNames)
	if err != nil {
		log.Fatal(err)
	}
}

type processor func(pr wiki.PageRepository, pages []wiki.Page) error

func process(pr wiki.PageRepository, batchSize int, f processor) error {
	pageBuffer, err := pr.NextPages("", batchSize)
	if err != nil {
		return err
	}

	for len(pageBuffer) != 0 {
		err = f(pr, pageBuffer)
		if err != nil {
			return err
		}

		pageBuffer, err = pr.NextPages(pageBuffer[len(pageBuffer)-1].Title, batchSize)
	}

	return nil
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
	return errors.New("not implemented")
}
