package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

const apiBaseUrl = "https://en.wikipedia.org/w/api.php"
const pageUrl = apiBaseUrl + "?action=query&prop=revisions&rvprop=content&format=json&titles="

func LoadPageContent(title string) (page Page, err error) {
	var body []byte
	body, err = loadPageContentFromApi(title)
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

func loadPageContentFromApi(title string) (body []byte, err error) {
	url := pageUrl + title
	response, err := http.Get(url)
	if err != nil {
		return
	}

	body, err = ioutil.ReadAll(response.Body)
	return
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
