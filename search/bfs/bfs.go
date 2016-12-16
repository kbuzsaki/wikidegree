/*
Implements a "Six degrees of separation" style search for the shortest "path"
between two wikipedia pages using a breadth first search.

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

This algorithm is pretty friendly to parallelization, though network
inconsistencies can make results inconsistent if multiple shortest paths
exist.

Its primary weakness is that it's a big memory hog for most searches that
require more than ~5 hops.
Hopefully iddfs.go will help with that :)
*/
package bfs

import (
	"context"

	"github.com/kbuzsaki/wikidegree/wiki"
)

const defaultFrontierSize = 10 * 1000 * 1000
const defaultNumScraperThreads = 10

func GetBfsPathFinder(pageLoader wiki.PageLoader) wiki.PathFinder {
	pathFinder := bfsPathFinder{pageLoader, defaultFrontierSize, defaultNumScraperThreads, false}
	return &pathFinder
}

// Implements wiki.PathFinder
type bfsPathFinder struct {
	pageLoader        wiki.PageLoader
	frontierSize      int
	numScraperThreads int
	serial            bool
}

// Implements wiki.PathFinder.SetPageLoader()
func (bpf *bfsPathFinder) SetPageLoader(pageLoader wiki.PageLoader) {
	bpf.pageLoader = pageLoader
}

// Implements wiki.PathFinder.FindPath()
func (bpf *bfsPathFinder) FindPath(ctx context.Context, start, end string) (wiki.TitlePath, error) {
	if bpf.serial {
		return bpf.findNearestPathSerial(start, end)
	} else {
		return bpf.findNearestPathParallel(ctx, start, end)
	}
}
