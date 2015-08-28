package main

import "encoding/json"
import "io/ioutil"
import "net/http"
import "regexp"
import "strings"

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


