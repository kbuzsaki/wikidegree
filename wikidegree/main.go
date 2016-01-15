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
	source string
	algorithm string
	start string
	end string
}

func main() {
	params, err := getParameters()
	if err != nil {
		log.Fatal(err)
	}

	pageLoader := getPageLoader(params.source)
	defer pageLoader.Close()

	pathFinder := getPathFinder(params.algorithm, pageLoader)

	// validate the start page
	startPage, err := pageLoader.LoadPage(params.start)
	if err != nil {
		log.Fatal("Start page '" + params.start + "' does not exist!")
	}
	if len(startPage.Links) == 0 {
		log.Fatal("Start page '" + params.start + "' has no links!")
	}

	// validate the end page
	_, err = pageLoader.LoadPage(params.end)
	if err != nil {
		log.Fatal("End page '" + params.end + "' does not exist!")
	}

	// actually perform search
	fmt.Println("Finding shortest path from", params.start, "to", params.end, "using", params.algorithm)
	fmt.Println()

	path, err := pathFinder.FindPath(params.start, params.end)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println()
	fmt.Println("Final path:", path)
}

func getParameters() (parameters, error) {
	sourcePtr := flag.String("src", "bolt", "the source for page loading")
	algorithmPtr := flag.String("alg", "bfs", "the path finding algorithm")
	flag.Parse()

	if flag.NArg() != 2 {
		return parameters{}, fmt.Errorf("Expected exactly 2 arguments (start and end), found %d", flag.NArg())
	}
	args := flag.Args()
	start := api.EncodeTitle(args[0])
	end := api.EncodeTitle(args[1])

	return parameters{*sourcePtr, *algorithmPtr, start, end}, nil
}

func getPageLoader(source string) api.PageLoader {
	switch source {
	case "bolt":
		pageLoader, err := api.GetBoltPageLoader()
		if err != nil {
			log.Fatal(err)
		}
		return pageLoader
	case "web":
		return api.GetWebPageLoader()
	default:
		log.Fatal("Unknown source:", source)
		return nil
	}
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

