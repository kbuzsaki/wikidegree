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
