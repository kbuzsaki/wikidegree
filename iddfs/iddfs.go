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
	api "github.com/kbuzsaki/wikidegree/api"
)

const MaxWorkerThreads = 10

const MaxDepth = 4

func FindNearestPathParallel(start string, end string) api.TitlePath {
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
	return depthLimitedSearchParallel(start, end, MaxDepth)
}

func depthLimitedSearchParallel(start string, end string, depthLimit int) api.TitlePath {
	requestQueue := make(chan chan<- api.TitlePath)
	loadedQueue := make(chan api.TitlePath)
	toLoadQueue := make(chan api.TitlePath)

	go DfsQueue(toLoadQueue, requestQueue)

	for i := 0; i < MaxWorkerThreads; i++ {
		go requestQueueWorker(requestQueue, loadedQueue)
	}

	toLoadQueue <- api.TitlePath{start}

	for titlePath := range loadedQueue {
		if titlePath.Head() == end {
			return titlePath
		} else if len(titlePath) <= depthLimit {
			toLoadQueue <- titlePath
		}
	}

	return nil
}

func requestQueueWorker(requestQueue chan<- chan<- api.TitlePath, output chan<- api.TitlePath) {
	input := make(chan api.TitlePath)

	for {
		requestQueue <- input
		titlePath := <-input

		fmt.Println("Loading:", titlePath)
		page, _ := api.LoadPageContent(titlePath.Head())
		parsedPage := api.ParsePage(page)

		for _, link := range parsedPage.Links {
			newTitlePath := titlePath.Catted(link)
			output <- newTitlePath
		}
	}
}

func FindNearestPathSerial(start string, end string) api.TitlePath {
	for depthLimit := 1; depthLimit <= MaxDepth; depthLimit++ {
		fmt.Println()
		fmt.Println("Beginning search with depth limit", depthLimit)
		path := depthLimitedSearchSerial(start, end, depthLimit)

		if path != nil {
			return path
		}
	}
	return nil
}

func depthLimitedSearchSerial(start string, end string, depthLimit int) api.TitlePath {
	// technically this isn't needed anymore,
	// since a depth limited search will handle cycles gracefully
	visited := make(map[string]bool)
	visited[start] = true

	var titlePath api.TitlePath
	titlePathStack := []api.TitlePath{{start}}

	for len(titlePathStack) > 0 {
		titlePath, titlePathStack = titlePathStack[len(titlePathStack)-1], titlePathStack[:len(titlePathStack)-1]

		fmt.Println("Loading:", titlePath)
		page, _ := api.LoadPageContent(titlePath.Head())
		parsedPage := api.ParsePage(page)

		for _, link := range parsedPage.Links {
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

