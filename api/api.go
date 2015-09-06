/*
Implements a simple interface for loading the content from wikipedia pages.
This lets you easily find all of the pages that a given wikipedia page
links to.

Currently loads data a bit messily using the json api.

In the future, an alternate implementation could load the data from a local
copy of wikipedia, even extending the interface to allow searching for all
links to a particular page.
 */
package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

const apiBaseUrl = "https://en.wikipedia.org/w/api.php"
const pageUrl = apiBaseUrl + "?action=query&prop=revisions&rvprop=content&format=json&titles="

type Page struct {
	Title   string
	Content string
}

type ParsedPage struct {
	Title string
	Links []string
}

func LoadPageContent(title string) (page Page, err error) {
	url := pageUrl + title
	response, err := http.Get(url)
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	var query jsonPageQuery
	err = json.Unmarshal(body, &query)
	if err != nil {
		return
	}

	for _, jsonPage := range query.Query.Pages {
		for _, revision := range jsonPage.Revisions {
			page = Page{title, revision["*"]}
			return
		}
	}
	return
}

func ParsePage(page Page) ParsedPage {
	regex, _ := regexp.Compile("\\[\\[(.+?)(\\]\\]|\\||#)")

	matches := regex.FindAllStringSubmatch(page.Content, -1)

	var links []string
	for _, match := range matches {
		link := match[1]
		link = encodeTitle(link)
		links = append(links, link)
	}

	return ParsedPage{page.Title, links}
}

func encodeTitle(title string) string {
	title = strings.Replace(title, " ", "_", -1)
	title = strings.ToLower(title)
	return title
}

type jsonPage struct {
	Pageid    int
	Title     string
	Revisions []map[string]string
}

type jsonPageQuery struct {
	Query struct {
		Pages map[string]jsonPage
	}
}
