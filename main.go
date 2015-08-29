package main

import "fmt"

func main() {
	fmt.Printf("Hello, world.\n")

    path := FindNearestPathDualGoroutine("hydrogen", "hungary")

    fmt.Printf("Final path:\n")
    for _, element := range path {
        fmt.Printf("%s\n", element)
    }
}

const FrontierSize = 10 * 1000 * 1000
const NumScraperThreads = 10

func FindNearestPathDualGoroutine(start string, end string) []string {
    titles := make(chan string, FrontierSize)
    pages := make(chan Page, 10)
    parsedPages := make(chan ParsedPage, 10)

    for i := 0; i < NumScraperThreads; i++ {
        go loadPages(titles, pages)
    }
    go parsePages(pages, parsedPages)

    titles <- start
    visited := make(map[string]string)
    visited[start] = ""

    for parsedPage := range parsedPages {
        for _, link := range parsedPage.links {
            if link == end {
                fmt.Printf("Done!\n\n")
                visited[link] = parsedPage.title
                return pathFromVisited(visited, start, end)
            } else if len(visited[link]) == 0 {
                visited[link] = parsedPage.title
                titles <- link
            }
        }
    }

    return nil
}

type Page struct {
    title string
    content string
}

type ParsedPage struct {
    title string
    links []string
}

func loadPages(titles <-chan string, pages chan<- Page) {
    for title := range titles {
        fmt.Printf("Loading: %s\n", title)
        if content, ok := LoadPageContent(title); ok {
            pages <- Page{title, content}
        } else {
            fmt.Printf("Failed to load '%s'\n", title)
        }
    }
}

func parsePages(pages <-chan Page, parsedPages chan<- ParsedPage) {
    for page := range pages {
        links := ParseTitles(page.content)
        parsedPages <- ParsedPage{page.title, links}
    }
}

func FindNearestPathSerial(start string, end string) []string {
    visited := make(map[string]string)
    visited[start] = ""
    frontier := []string{start}

    for len(frontier) > 0 {
        if len(visited[end]) > 0 {
            fmt.Printf("Done!\n\n")
            return pathFromVisited(visited, start, end)
        }

        var next string
        next, frontier = frontier[0], frontier[1:]

        fmt.Printf("Loading: %s\n", next)
        if content, ok := LoadPageContent(next); ok {
            titles := ParseTitles(content)

            for _, title := range titles {
                if len(visited[title]) == 0 {
                    frontier = append(frontier, title)
                    visited[title] = next
                }
            }
        } else {
            fmt.Printf("Failed to load '%s'\n", next)
        }
    }

    return nil
}

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
