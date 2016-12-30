package consumers

import (
	"log"
	"sync"

	"github.com/kbuzsaki/wikidegree/batch"
	"github.com/kbuzsaki/wikidegree/wiki"
)

func SavePageBuffers(wg *sync.WaitGroup, config batch.Config, ps wiki.PageSaver, pageBuffers <-chan []wiki.Page) {
	defer wg.Done()

	counter := 0
	for pageBuffer := range pageBuffers {
		err := ps.SavePages(pageBuffer)
		if err != nil {
			log.Fatal(err)
		}

		counter += len(pageBuffer)
		if config.Debug {
			log.Println("saved", counter)
		}
	}
}
