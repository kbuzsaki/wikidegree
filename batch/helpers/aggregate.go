package helpers

import "github.com/kbuzsaki/wikidegree/wiki"

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

func AggregatePageBlobs(bufferSize int, pages <-chan wiki.Page, pageBuffers chan<- []wiki.Page) {
	var pageBuffer []wiki.Page
	pageLookup := make(map[string]*wiki.Page)

	for page := range pages {
		// check if there's already a buffered entry for this page, if there is then just merge the blobs
		if currPage, ok := pageLookup[page.Title]; ok {
			for key, val := range page.Blob {
				currPage.SetBlob(key, val)
			}
		} else {
			pageBuffer = append(pageBuffer, page)
			pageLookup[page.Title] = &pageBuffer[len(pageBuffer) - 1]

			if len(pageBuffer) >= bufferSize {
				pageBuffers <- pageBuffer
				pageBuffer = nil
				pageLookup = make(map[string]*wiki.Page)
			}
		}
	}

	close(pageBuffers)
}
