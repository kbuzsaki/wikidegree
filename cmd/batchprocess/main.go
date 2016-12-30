package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"sync"

	"github.com/kbuzsaki/wikidegree/batch"
	"github.com/kbuzsaki/wikidegree/batch/processors"
	"github.com/kbuzsaki/wikidegree/wiki"
	"github.com/kbuzsaki/wikidegree/batch/consumers"
	"github.com/kbuzsaki/wikidegree/batch/helpers"
)

const saveBufferSize = 10000

const defaultBatchSize = 10000
const defaultConcurrency = 1

func main() {
	debug := flag.Bool("debug", false, "print debug output")
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

	config := batch.Config{
		BatchSize:   *batchSize,
		Concurrency: *concurrency,
		Debug:       *debug,
	}

	doFilterDeadLinks(config, pr)
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
	wg := &sync.WaitGroup{}
	pages := make(chan wiki.Page, (config.BatchSize*config.Concurrency)/2)
	pageBuffers := make(chan []wiki.Page, 2*config.Concurrency)

	processor, err := processors.NewBlobReverseLinker(config, pr, pages)
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
