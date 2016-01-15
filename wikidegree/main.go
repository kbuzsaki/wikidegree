/*
Simple main package for running the different search algorithms.
Maybe this will eventually turn into a decent command line interface.
*/
package main

import (
	"fmt"
	"log"
	bfs "github.com/kbuzsaki/wikidegree/bfs"
	iddfs "github.com/kbuzsaki/wikidegree/iddfs"
	api "github.com/kbuzsaki/wikidegree/api"
)


func main() {
	algorithm := "bfs"

	pageLoader := api.GetWebPageLoader()

	var pathFinder api.PathFinder
	switch algorithm {
	case "bfs":
		pathFinder = bfs.GetBfsPathFinder(pageLoader)
	case "iddfs":
		pathFinder = iddfs.GetIddfsPathFinder(pageLoader)
	}

	start := "hydrogen"
	end := "hungary"

	fmt.Println("Finding shortest path from", start, "to", end, "using", algorithm)
	fmt.Println()

	path, err := pathFinder.FindPath(start, end)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Final path:", path)
}

