package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"sync"

	"github.com/kbuzsaki/wikidegree/batch"
	"github.com/kbuzsaki/wikidegree/batch/consumers"
	"github.com/kbuzsaki/wikidegree/batch/helpers"
	"github.com/kbuzsaki/wikidegree/batch/processors"
	"github.com/kbuzsaki/wikidegree/wiki"
)

const saveBufferSize = 750000

const defaultBatchSize = 10000

// 6 has the highest throughput for the pageNopper processor
const defaultConcurrency = 6

const newDbFilename = "db/new.db"

func main() {
	debug := flag.Bool("debug", false, "print debug output")
	dbFilename := flag.String("db", wiki.DefaultIndexName, "the boltdb file")
	batchSize := flag.Int("batch", defaultBatchSize, "number of pages to pass to the processing function at a time")
	concurrency := flag.Int("conc", defaultConcurrency, "number of goroutines to use for processing")
	skip := flag.Int("skip", 0, "number of titles to skip initially")
	flag.Parse()

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	pr, err := wiki.GetBoltPageRepository(*dbFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer pr.Close()

	config := batch.Config{
		BatchSize:   *batchSize,
		Concurrency: *concurrency,
		Skip:        *skip,
		Debug:       *debug,
	}

	//doFilterDeadPages(config, pr)
	//doFilterDeadLinks(config, pr)
	doBlobReverseLinks(config, pr)

}

func doFilterDeadPages(config batch.Config, pr wiki.PageRepository) {
	wg := &sync.WaitGroup{}
	deadTitles := make(chan string, (config.BatchSize*config.Concurrency)/2)

	processor, err := processors.NewDeadTitleFilterer(config, pr, deadTitles)
	if err != nil {
		log.Fatal(err)
	}

	go consumers.DeleteTitles(wg, config, pr, deadTitles)
	wg.Add(1)

	err = batch.RunTitleJob(pr, processor, config)
	if err != nil {
		log.Fatal("error running batch job: ", err)
	}

	wg.Wait()
}

func doFilterDeadLinks(config batch.Config, pr wiki.PageRepository) {
	wg := &sync.WaitGroup{}
	pages := make(chan wiki.Page, (config.BatchSize*config.Concurrency)/2)
	pageBuffers := make(chan []wiki.Page, 2*config.Concurrency)

	processor, err := processors.NewDeadLinkFilterer(config, pr, pages)
	if err != nil {
		log.Fatal(err)
	}

	go helpers.AggregatePages(saveBufferSize, pages, pageBuffers)
	go consumers.SavePageBuffers(wg, config, pr, pageBuffers)
	wg.Add(1)

	err = batch.RunPageJob(pr, processor, config)
	if err != nil {
		log.Fatal("error running batch job: ", err)
	}

	wg.Wait()
}

func doBlobReverseLinks(config batch.Config, pr wiki.PageRepository) {
	outPr, err := wiki.GetBoltPageRepository(newDbFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer outPr.Close()

	predicates, err := getTitleRangePredicates(pr, saveBufferSize)
	if err != nil {
		log.Fatal(err)
	}

	for _, predicate := range predicates {
		wg := &sync.WaitGroup{}
		pages := make(chan wiki.Page, (config.BatchSize*config.Concurrency)/2)
		pageBuffers := make(chan []wiki.Page)
		chunkedPageBuffers := make(chan []wiki.Page)

		/*
			processor, err := processors.NewFilteringBlobReverseLinker(config, predicate, pages)
			if err != nil {
				log.Fatal(err)
			}

			go helpers.AggregatePageBlobs(pages, pageBuffers)
		*/

		processor, err := processors.NewFilteringReverseLinker(config, predicate, pages)
		if err != nil {
			log.Fatal(err)
		}

		go helpers.AggregatePageLinkers(pages, pageBuffers)
		go helpers.ChunkPageBuffers(10000, pageBuffers, chunkedPageBuffers)
		go consumers.SavePageBuffers(wg, config, outPr, chunkedPageBuffers)
		wg.Add(1)

		err = batch.RunPageJob(pr, processor, config)
		if err != nil {
			log.Fatal("error running batch job: ", err)
		}

		wg.Wait()
	}
}

func getTitleRangePredicates(pr wiki.PageRepository, skipSize int) ([]helpers.PagePredicate, error) {
	var predicates []helpers.PagePredicate

	startTitle := ""
	for {
		endTitle, err := pr.SkipTitles(startTitle, skipSize)
		if err != nil {
			return nil, err
		}

		predicates = append(predicates, makeRangePredicate(startTitle, endTitle))

		// if endTitle is empty string, then we've hit the end
		if endTitle == "" {
			return predicates, nil
		}

		startTitle = endTitle
	}
}

func makeRangePredicate(startTitle, endTitle string) helpers.PagePredicate {
	return func(page wiki.Page) bool {
		return (startTitle <= page.Title) && (page.Title < endTitle || endTitle == "")
	}
}

func doCountLinks(config batch.Config, pr wiki.PageRepository) {
	wg := &sync.WaitGroup{}
	linkCounts := make(chan int, 2*config.Concurrency)

	processor, err := processors.NewLinkCounter(config, linkCounts)
	if err != nil {
		log.Fatal(err)
	}

	go consumers.HistogramInts(wg, config, linkCounts)
	wg.Add(1)

	err = batch.RunPageJob(pr, processor, config)
	if err != nil {
		log.Fatal("error running batch job: ", err)
	}

	wg.Wait()
}

func doTitleNopper(config batch.Config, pr wiki.PageRepository) {
	processor, err := processors.NewTitleNopper(config)
	if err != nil {
		log.Fatal(err)
	}

	err = batch.RunTitleJob(pr, processor, config)
	if err != nil {
		log.Fatal(err)
	}
}

func doPageNopper(config batch.Config, pr wiki.PageRepository) {
	processor, err := processors.NewPageNopper(config)
	if err != nil {
		log.Fatal(err)
	}

	err = batch.RunPageJob(pr, processor, config)
	if err != nil {
		log.Fatal(err)
	}
}
