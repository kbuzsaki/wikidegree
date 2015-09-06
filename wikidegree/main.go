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
	fmt.Printf("Hello, world.\n")

	path := bfs.FindNearestPathParallel("hydrogen", "hungary")

	fmt.Printf("Final path:\n")
	for _, element := range path {
		fmt.Printf("%s\n", element)
	}
}
