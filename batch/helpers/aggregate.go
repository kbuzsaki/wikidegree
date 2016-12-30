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
