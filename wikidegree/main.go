/*
Simple main package for running the different search algorithms.
Maybe this will eventually turn into a decent command line interface.
*/
package main

import (
	"fmt"
	bfs "github.com/kbuzsaki/wikidegree/bfs"
)

func main() {
	start := "hydrogen"
	end := "hungary"

	fmt.Println("Finding shortest path from", start, "to", end)
	fmt.Println()

	path := bfs.FindNearestPathParallel(start, end)
	fmt.Println("Final path:", path)
}
