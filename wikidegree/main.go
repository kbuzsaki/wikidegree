/*
Simple main package for running the different search algorithms.
Maybe this will eventually turn into a decent command line interface.
*/
package main

import (
	"fmt"
	bfs "github.com/kbuzsaki/wikidegree/bfs"
	iddfs "github.com/kbuzsaki/wikidegree/iddfs"
)


func main() {
	algorithm := "bfs"

	start := "hydrogen"
	end := "hungary"

	fmt.Println("Finding shortest path from", start, "to", end, "using", algorithm)
	fmt.Println()

	var path []string

	switch algorithm {
	case "bfs":
		path = bfs.FindNearestPathParallel(start, end)
	case "iddfs":
		path = iddfs.FindNearestPathSerial(start, end)
	}

	fmt.Println("Final path:", path)
}
