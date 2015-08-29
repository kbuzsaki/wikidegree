package main

import "encoding/json"
import "io/ioutil"
import "net/http"
import "regexp"
import "strings"

const WikiApiBase = "https://en.wikipedia.org/w/api.php"
const WikiPageUrl = WikiApiBase + "?action=query&prop=revisions&rvprop=content&format=json&titles="

type WikiPage struct {
    title string
    content string
}

type ParsedWikiPage struct {
    title string
    links []string
}

func LoadPageContent(title string) (page WikiPage, err error) {
    url := WikiPageUrl + title
    response, err := http.Get(url)
    if err != nil {
        return
    }

    body, err := ioutil.ReadAll(response.Body)
    if err != nil {
        return
    }

    var query jsonWikiPageQuery
    err = json.Unmarshal(body, &query)
    if err != nil {
        return
    }

    for _, jsonPage := range query.Query.Pages {
        for _, revision := range jsonPage.Revisions {
            page = WikiPage{title, revision["*"]}
            return
        }
    }
    return
}

func ParsePage(page WikiPage) ParsedWikiPage {
    regex, _ := regexp.Compile("\\[\\[(.+?)(\\]\\]|\\||#)")

    matches := regex.FindAllStringSubmatch(page.content, -1)

    var links []string
    for _, match := range matches {
        link := match[1]
        link = encodeTitle(link)
        links = append(links, link)
    }

    return ParsedWikiPage{page.title, links}
}

func encodeTitle(title string) string {
    title = strings.Replace(title, " ", "_", -1)
    title = strings.ToLower(title)
    return title
}

type jsonWikiPage struct {
    Pageid int
    Title string
    Revisions []map[string]string
}

type jsonWikiPageQuery struct {
    Query struct {
        Pages map[string]jsonWikiPage
    }
}

