package batch

import (
	"log"
	_ "net/http/pprof"
	"sync"

	"github.com/kbuzsaki/wikidegree/wiki"
)

const printThresh = 10000

type PageProcessor interface {
	Setup() error
	Process(page wiki.Page) error
	Teardown() error
}

type Config struct {
	BatchSize   int
	Concurrency int
	Debug       bool
}

func RunJob(pr wiki.PageRepository, processor PageProcessor, config Config) error {
	wg := &sync.WaitGroup{}
	titleBuffers := make(chan []string, 2*config.Concurrency)
	errs := make(chan error)

	err := processor.Setup()
	if err != nil {
		return err
	}

	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go jobWorker(wg, pr, processor, titleBuffers, errs)
	}

	err = runJob(pr, config, titleBuffers, errs)
	if err != nil {
		close(titleBuffers)
		close(errs)
		return err
	}

	wg.Wait()

	err = processor.Teardown()
	if err != nil {
		return err
	}

	return nil
}

func runJob(pr wiki.PageRepository, config Config, titleBuffers chan<- []string, errs <-chan error) error {
	titleBuffer, err := pr.NextTitles("", config.BatchSize)
	if err != nil {
		return err
	}

	counter := 0
	for len(titleBuffer) != 0 {
		counter += len(titleBuffer)
		if config.Debug && counter%printThresh == 0 {
			log.Println("processed", counter)
		}

		select {
		case titleBuffers <- titleBuffer:
			titleBuffer, err = pr.NextTitles(titleBuffer[len(titleBuffer)-1], config.BatchSize)
		case err := <-errs:
			return err
		}
	}

	return nil
}

func jobWorker(wg *sync.WaitGroup, pr wiki.PageRepository, processor PageProcessor, titleBuffers <-chan []string, errs chan<- error) {
	defer wg.Done()

	for titleBuffer := range titleBuffers {
		pageBuffer, err := pr.LoadPages(titleBuffer)
		if err != nil {
			errs <- err
			return
		}

		for _, page := range pageBuffer {
			err := processor.Process(page)
			if err != nil {
				errs <- err
				return
			}
		}
	}
}
