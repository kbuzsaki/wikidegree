package main

import "fmt"

func main() {
	fmt.Printf("Hello, world.\n")

    path := FindNearestPathBfsParallel("hydrogen", "hungary")

    fmt.Printf("Final path:\n")
    for _, element := range path {
        fmt.Printf("%s\n", element)
    }
}

