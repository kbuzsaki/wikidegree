package helpers

import (
	"log"

	"github.com/kbuzsaki/wikidegree/wiki"
)

const printThresh = 1000000

func AggregatePages(bufferSize int, pages <-chan wiki.Page, pageBuffers chan<- []wiki.Page) {
	var pageBuffer []wiki.Page

	for page := range pages {
		pageBuffer = append(pageBuffer, page)

		if len(pageBuffer) >= bufferSize {
			pageBuffers <- pageBuffer
			pageBuffer = nil
		}
	}

	close(pageBuffers)
}

func AggregatePageBlobs(pages <-chan wiki.Page, pageBuffers chan<- []wiki.Page) {
	var pageBuffer []wiki.Page
	pageLookup := make(map[string]*wiki.Page)
	total := 0
	counter := 0

	for page := range pages {
		total++
		counter++
		if counter%printThresh == 0 {
			compression := float64(len(pageBuffer))/float64(counter)
			backlog := float64(len(pages))/float64(cap(pages))
			log.Printf("aggregate blobs: buffer=%d, counter=%d, compression=%0.3f, total=%d, backlog=%0.3f\n", len(pageBuffer), counter, compression, total, backlog)
		}

		select {
		case pageBuffers <- pageBuffer:
			log.Printf("sent buffer with length %d\n", len(pageBuffer))
			pageBuffer = nil
			pageLookup = make(map[string]*wiki.Page)
			counter = 0
		default:
			// check if there's already a buffered entry for this page, if there is then just merge the blobs
			if currPage, ok := pageLookup[page.Title]; ok {
				for key, val := range page.Blob {
					currPage.SetBlob(key, val)
				}
			} else {
				pageBuffer = append(pageBuffer, page)
				pageLookup[page.Title] = &pageBuffer[len(pageBuffer)-1]
			}
		}
	}

	close(pageBuffers)
}
