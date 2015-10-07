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

// pseudo tuple containing the depth of this page in the explored tree
type DepthTitle struct {
	Depth int
	Title string
}

const MaxDepth = 4

func FindNearestPathSerial(start string, end string) []string {
	for depthLimit := 1; depthLimit < MaxDepth; depthLimit++ {
		fmt.Println()
		fmt.Println("Beginning search with depth limit", depthLimit)
		path := depthLimitedSearchSerial(start, end, depthLimit)

		if path != nil {
			return path
		}
	}
	return nil
}

func depthLimitedSearchSerial(start string, end string, depthLimit int) []string {
	visited := make(map[string]string)
	visited[start] = ""

	var depthTitle DepthTitle
	titleStack := []DepthTitle{{0, start}}

	for len(titleStack) > 0 {
		depthTitle, titleStack = titleStack[len(titleStack)-1], titleStack[:len(titleStack)-1]

		fmt.Println("Loading:", depthTitle)
		page, _ := api.LoadPageContent(depthTitle.Title)
		parsedPage := api.ParsePage(page)

		for _, link := range parsedPage.Links {
			if link == end {
				fmt.Println("Done!")
				fmt.Println()
				visited[link] = depthTitle.Title
				return pathFromVisited(visited, start, end)
			} else if depthTitle.Depth < depthLimit && len(visited[link]) == 0 {
				visited[link] = depthTitle.Title
				linkDepthTitle := DepthTitle{depthTitle.Depth + 1, link}
				titleStack = append(titleStack, linkDepthTitle)
			}
		}
	}

	return nil
}


// this requires linear memory with respect to the number of pages visited...
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
