package main

import "encoding/json"
import "fmt"
import "io/ioutil"
import "net/http"
import "regexp"
import "strings"

func main() {
	fmt.Printf("Hello, world.\n")

    path := FindNearestPath("hydrogen", "hungary")

    fmt.Printf("Final path:\n")
    for _, element := range path {
        fmt.Printf("%s\n", element)
    }
}

func FindNearestPath(start string, end string) []string {
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

const WikiApiBase = "https://en.wikipedia.org/w/api.php"
const WikiPageUrl = WikiApiBase + "?action=query&prop=revisions&rvprop=content&format=json&titles="

type WikiPage struct {
    Pageid int
    Title string
    Revisions []map[string]string
}

type WikiPageQuery struct {
    Query struct {
        Pages map[string]WikiPage
    }
}

func LoadPageContent(title string) (string, bool) {
    url := WikiPageUrl + title
    response, _ := http.Get(url)
    body, _ := ioutil.ReadAll(response.Body)

    var query WikiPageQuery
    json.Unmarshal(body, &query)

    for _, page := range query.Query.Pages {
        for _, revision := range page.Revisions {
            return revision["*"], true
        }
    }
    return "", false
}

func ParseTitles(content string) []string {
    regex, _ := regexp.Compile("\\[\\[(.+?)(\\]\\]|\\||#)")

    matches := regex.FindAllStringSubmatch(content, -1)

    var titles []string
    for _, match := range matches {
        title := match[1]
        title = encodeTitle(title)
        titles = append(titles, title)
    }

    return titles
}

func encodeTitle(title string) string {
    title = strings.Replace(title, " ", "_", -1)
    title = strings.ToLower(title)
    return title
}

