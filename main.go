package main

import "encoding/json"
import "fmt"
import "io/ioutil"
import "net/http"
import "regexp"
import "strings"

func main() {
	fmt.Printf("Hello, world.\n")

    initialTitle := "hydrogen"
    content, ok := LoadPageContent(initialTitle)

    if ok {
        fmt.Printf("\ncontent:\n%s\n", content)
        titles := ParseTitles(content)
        fmt.Printf("\ntitles:\n")
        for i, title := range titles {
            fmt.Printf("%d: %s\n", i, title)
        }
    } else {
        fmt.Printf("failure?\n")
    }
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

