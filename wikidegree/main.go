/*
Simple main package for running the different search algorithms.
Maybe this will eventually turn into a decent command line interface.
*/
package main

import (
	"flag"
	"fmt"
	"log"
	bfs "github.com/kbuzsaki/wikidegree/bfs"
	iddfs "github.com/kbuzsaki/wikidegree/iddfs"
	api "github.com/kbuzsaki/wikidegree/api"
)

type parameters struct {
	algorithm string
	start string
	end string
}

func main() {
	params, err := getParameters()
	if err != nil {
		log.Fatal(err)
	}

	pageLoader := getPageLoader()
	pathFinder := getPathFinder(params.algorithm, pageLoader)

	fmt.Println("Finding shortest path from", params.start, "to", params.end, "using", params.algorithm)
	fmt.Println()

	path, err := pathFinder.FindPath(params.start, params.end)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Final path:", path)
}

func getParameters() (parameters, error) {
	algorithmPtr := flag.String("alg", "bfs", "the path finding algorithm")
	flag.Parse()

	if flag.NArg() != 2 {
		return parameters{}, fmt.Errorf("Expected exactly 2 arguments (start and end), found %d", flag.NArg())
	}
	args := flag.Args()
	start := args[0]
	end := args[1]

	return parameters{*algorithmPtr, start, end}, nil
}

func getPageLoader() api.PageLoader {
	return api.GetWebPageLoader()
}

func getPathFinder(algorithm string, pageLoader api.PageLoader) api.PathFinder {
	switch algorithm {
	case "bfs":
		return bfs.GetBfsPathFinder(pageLoader)
	case "iddfs":
		return iddfs.GetIddfsPathFinder(pageLoader)
	default:
		log.Fatal("Unknown path finding algorithm: ", algorithm)
		return nil
	}
}

