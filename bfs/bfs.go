package bfs

import "fmt"
import api "github.com/kbuzsaki/wikidegree/api"

const frontierSize = 10 * 1000 * 1000
const numScraperThreads = 10

func FindNearestPathParallel(start string, end string) []string {
    titles := make(chan string, frontierSize)
    pages := make(chan api.WikiPage, 10)
    parsedPages := make(chan api.ParsedWikiPage, 10)

    for i := 0; i < numScraperThreads; i++ {
        go loadPages(titles, pages)
    }
    go parsePages(pages, parsedPages)

    titles <- start
    visited := make(map[string]string)
    visited[start] = ""

    for parsedPage := range parsedPages {
        for _, link := range parsedPage.Links {
            if link == end {
                fmt.Printf("Done!\n\n")
                visited[link] = parsedPage.Title
                return pathFromVisited(visited, start, end)
            } else if len(visited[link]) == 0 {
                visited[link] = parsedPage.Title
                titles <- link
            }
        }
    }

    return nil
}

func loadPages(titles <-chan string, pages chan<- api.WikiPage) {
    for title := range titles {
        fmt.Printf("Loading: %s\n", title)
        if page, err := api.LoadPageContent(title); err == nil {
            pages <- page
        } else {
            fmt.Printf("Failed to load '%s'\n", title)
        }
    }
}

func parsePages(pages <-chan api.WikiPage, parsedPages chan<- api.ParsedWikiPage) {
    for page := range pages {
        parsedPages <- api.ParsePage(page)
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
        if page, err := api.LoadPageContent(next); err == nil {
            parsedPage := api.ParsePage(page)

            for _, title := range parsedPage.Links {
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
