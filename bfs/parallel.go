package bfs

import (
	"errors"
	"log"
	api "github.com/kbuzsaki/wikidegree/api"
)

// parallel implementation of bfs
func (bpf* bfsPathFinder) findNearestPathParallel(start, end string) (api.TitlePath, error) {
	titles := make(chan string, bpf.frontierSize)
	pages := make(chan api.Page)

	for i := 0; i < bpf.numScraperThreads; i++ {
		go bpf.loadPages(titles, pages)
	}

	titles <- start
	visited := make(map[string]string)
	visited[start] = ""

	for page := range pages {
		for _, link := range page.Links {
			if link == end {
				// close the channels to halt other goroutines
				close(titles)

				log.Println("Found end page:", end, "stopping...")
				visited[link] = page.Title
				return pathFromVisited(visited, start, end), nil
			} else if len(visited[link]) == 0 {
				visited[link] = page.Title
				titles <- link
			}
		}
	}

	return nil, errors.New("Ran out of links!")
}

// simple function for loading pages from the loader
func (bpf* bfsPathFinder) loadPages(titles <-chan string, pages chan<- api.Page) {
	for title := range titles {
		log.Println("Loading page:", title)
		if page, err := bpf.pageLoader.LoadPage(title); err == nil {
			pages <- page
		} else {
			log.Println("Error loading page:", title, "error:", err)
		}
	}
}

func pathFromVisited(visited map[string]string, start string, end string) []string {
	// starts from the end of the graph and pops back
	var path []string
	parent := end
	for parent != start {
		path = append(path, parent)
		parent = visited[parent]
	}
	path = append(path, start)

	// reverse the path before returning
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return path
}
