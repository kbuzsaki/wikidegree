package batch

import (
	"log"
	_ "net/http/pprof"
	"sync"

	"github.com/kbuzsaki/wikidegree/wiki"
)

const printThresh = 10000

type TitleProcessor interface {
	Setup() error
	ProcessTitle(title string) error
	Teardown() error
}

type PageProcessor interface {
	Setup() error
	ProcessPage(page wiki.Page) error
	Teardown() error
}

type Config struct {
	BatchSize   int
	Concurrency int
	Debug       bool
}

func RunTitleJob(pr wiki.PageRepository, processor TitleProcessor, config Config) error {
	wg, titleBuffers, errs, err := doSetup(config, processor.Setup)
	if err != nil {
		return err
	}

	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go titleJobWorker(wg, pr, processor, titleBuffers, errs)
	}

	err = runJob(pr, config, titleBuffers, errs)
	if err != nil {
		close(titleBuffers)
		return err
	}
	wg.Wait()

	err = processor.Teardown()
	if err != nil {
		return err
	}

	return nil
}

func RunPageJob(pr wiki.PageRepository, processor PageProcessor, config Config) error {
	wg, titleBuffers, errs, err := doSetup(config, processor.Setup)
	if err != nil {
		return err
	}

	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go pageJobWorker(wg, pr, processor, titleBuffers, errs)
	}

	err = runJob(pr, config, titleBuffers, errs)
	if err != nil {
		close(titleBuffers)
		return err
	}
	wg.Wait()

	err = processor.Teardown()
	if err != nil {
		return err
	}

	return nil
}

func doSetup(config Config, setup func() error) (*sync.WaitGroup, chan []string, chan error, error) {
	err := setup()
	if err != nil {
		return nil, nil, nil, err
	}

	wg := &sync.WaitGroup{}
	titleBuffers := make(chan []string, 2*config.Concurrency)
	errs := make(chan error, config.Concurrency)
	return wg, titleBuffers, errs, nil
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

func titleJobWorker(wg *sync.WaitGroup, pr wiki.PageRepository, processor TitleProcessor, titleBuffers <-chan []string, errs chan<- error) {
	defer wg.Done()

	for titleBuffer := range titleBuffers {
		for _, title := range titleBuffer {
			err := processor.ProcessTitle(title)
			if err != nil {
				log.Println("error processing title:", err)
				errs <- err
				return
			}
		}
	}
}

func pageJobWorker(wg *sync.WaitGroup, pr wiki.PageRepository, processor PageProcessor, titleBuffers <-chan []string, errs chan<- error) {
	defer wg.Done()

	for titleBuffer := range titleBuffers {
		pageBuffer, err := pr.LoadPages(titleBuffer)
		if err != nil {
			log.Println("error loading pages:", err)
			errs <- err
			return
		}

		for _, page := range pageBuffer {
			err := processor.ProcessPage(page)
			if err != nil {
				log.Println("error processing page:", err)
				errs <- err
				return
			}
		}
	}
}
