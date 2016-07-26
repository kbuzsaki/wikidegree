/*
Implements a "Six degrees of separation" style search for the shortest "path"
between two wikipedia pages using an iterative deepening depth first search.

The search treats the whole of wikipedia as a directed graph where each link
from a page A to another page B is treated as a directed edge from A to B.

For more info, see the wikipedia article:
https://en.wikipedia.org/wiki/Wikipedia:Six_degrees_of_Wikipedia

Here's an example of a (real!) shortest path from "bubble_gum" to "vladimir_putin"

bubble_gum
acetophenone
chemical_formula
nuclear_chemistry
vladimir_putin

This algorithm is less friendly to parallelization than the breadth first
search, but it should give much better memory efficiency at the cost of
additioanl cpu and network.
*/
package iddfs

import (
	"fmt"

	"github.com/kbuzsaki/wikidegree/wiki"
	"golang.org/x/net/context"
)

const defaultMaxWorkerThreads = 10
const defaultMaxDepth = 4

func GetIddfsPathFinder(pageLoader wiki.PageLoader) wiki.PathFinder {
	pathFinder := iddfsPathFinder{pageLoader, defaultMaxWorkerThreads, defaultMaxDepth, true}
	return &pathFinder
}

// Implements wiki.PathFinder
type iddfsPathFinder struct {
	pageLoader       wiki.PageLoader
	maxWorkerThreads int
	maxDepth         int
	serial           bool
}

// Implements wiki.PathFinder.SetPageLoader()
func (ipf *iddfsPathFinder) SetPageLoader(pageLoader wiki.PageLoader) {
	ipf.pageLoader = pageLoader
}

// Implements wiki.PathFinder.FindPath()
func (ipf *iddfsPathFinder) FindPath(ctx context.Context, start, end string) (wiki.TitlePath, error) {
	var path wiki.TitlePath

	if ipf.serial {
		path = ipf.findNearestPathSerial(start, end)
	} else {
		path = ipf.findNearestPathParallel(start, end)
	}

	return path, nil
}

func (ipf *iddfsPathFinder) findNearestPathParallel(start string, end string) wiki.TitlePath {
	// iterative deepening parallel currently doesn't work,
	// the program will deadlock once there is nothing left to read
	//
	// NOTE: that means this algorithm is *NOT* currently guaranteed
	//       to find the shortest path from start to end
	//
	/*
		for depthLimit := 2; depthLimit <= MaxDepth; depthLimit++ {
			fmt.Println()
			fmt.Println("Beginning search with depth limit", depthLimit)
			path := depthLimitedSearchParallel(start, end, depthLimit)

			if path != nil {
				return path
			}
		}
	*/

	// just do a regular depth limited search instead
	return ipf.depthLimitedSearchParallel(start, end, ipf.maxDepth)
}

func (ipf *iddfsPathFinder) depthLimitedSearchParallel(start string, end string, depthLimit int) wiki.TitlePath {
	requestQueue := make(chan chan<- wiki.TitlePath)
	loadedQueue := make(chan wiki.TitlePath)
	toLoadQueue := make(chan wiki.TitlePath)

	go DfsQueue(toLoadQueue, requestQueue)

	for i := 0; i < ipf.maxWorkerThreads; i++ {
		go ipf.requestQueueWorker(requestQueue, loadedQueue)
	}

	toLoadQueue <- wiki.TitlePath{start}

	for titlePath := range loadedQueue {
		if titlePath.Head() == end {
			return titlePath
		} else if len(titlePath) <= depthLimit {
			toLoadQueue <- titlePath
		}
	}

	return nil
}

func (ipf *iddfsPathFinder) requestQueueWorker(requestQueue chan<- chan<- wiki.TitlePath, output chan<- wiki.TitlePath) {
	input := make(chan wiki.TitlePath)

	for {
		requestQueue <- input
		titlePath := <-input

		fmt.Println("Loading:", titlePath)
		page, _ := ipf.pageLoader.LoadPage(titlePath.Head())

		for _, link := range page.Links {
			newTitlePath := titlePath.Catted(link)
			output <- newTitlePath
		}
	}
}

func (ipf *iddfsPathFinder) findNearestPathSerial(start string, end string) wiki.TitlePath {
	for depthLimit := 1; depthLimit <= ipf.maxDepth; depthLimit++ {
		fmt.Println()
		fmt.Println("Beginning search with depth limit", depthLimit)
		path := ipf.depthLimitedSearchSerial(start, end, depthLimit)

		if path != nil {
			return path
		}
	}
	return nil
}

func (ipf *iddfsPathFinder) depthLimitedSearchSerial(start string, end string, depthLimit int) wiki.TitlePath {
	// technically this isn't needed anymore,
	// since a depth limited search will handle cycles gracefully
	visited := make(map[string]bool)
	visited[start] = true

	var titlePath wiki.TitlePath
	titlePathStack := []wiki.TitlePath{{start}}

	for len(titlePathStack) > 0 {
		titlePath, titlePathStack = titlePathStack[len(titlePathStack)-1], titlePathStack[:len(titlePathStack)-1]

		fmt.Println("Loading:", titlePath)
		page, _ := ipf.pageLoader.LoadPage(titlePath.Head())

		for _, link := range page.Links {
			newTitlePath := titlePath.Catted(link)
			if link == end {
				fmt.Println("Done!")
				fmt.Println()
				return newTitlePath
			} else if len(newTitlePath) <= depthLimit && !visited[link] {
				visited[link] = true
				titlePathStack = append(titlePathStack, newTitlePath)
			}
		}
	}

	return nil
}
