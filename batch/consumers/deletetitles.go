package consumers

import (
	"log"
	"sync"

	"github.com/kbuzsaki/wikidegree/batch"
	"github.com/kbuzsaki/wikidegree/wiki"
)

func DeleteTitles(wg *sync.WaitGroup, config batch.Config, pr wiki.PageRepository, titles <-chan string) {
	defer wg.Done()

	counter := 0
	for title := range titles {
		err := pr.DeleteTitle(title)
		if err != nil {
			log.Fatal(err)
		}

		counter++
		if config.Debug {
			log.Printf("deleted title %#v, (%d)\n", title, counter)
		}
	}
}
