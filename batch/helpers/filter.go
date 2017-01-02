package helpers

import "github.com/kbuzsaki/wikidegree/wiki"

type PagePredicate func(page wiki.Page) bool

func FilterPages(predicate PagePredicate, pages <-chan wiki.Page, filteredPages chan<- wiki.Page) {
	for page := range pages {
		if predicate(page) {
			filteredPages <- page
		}
	}

	close(filteredPages)
}
