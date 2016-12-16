package wiki

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

const wikiBaseUrl = "https://en.wikipedia.org/w/api.php"
const defaultPageUrl = wikiBaseUrl + "?action=query&prop=revisions&rvprop=content&format=json&titles="

type webLoader struct {
	pageUrl string
}

func GetWebPageLoader() PageLoader {
	return webLoader{defaultPageUrl}
}

func (wl webLoader) LoadPage(title string) (Page, error) {
	body, err := wl.loadPageContentFromApi(title)
	if err != nil {
		return Page{}, err
	}

	var query jsonPageQuery
	err = json.Unmarshal(body, &query)
	if err != nil {
		return Page{}, err
	}

	for _, jsonPage := range query.Query.Pages {
		for _, revision := range jsonPage.Revisions {
			content := revision["*"]
			links := ParseLinks(content)

			// TODO: actually implement redirect support in the web loader
			redirector := title
			page := Page{redirector, title, links}
			return page, nil
		}
	}

	return Page{}, errors.New(fmt.Sprint("No revisions found for", title))
}

func (wl webLoader) Close() error {
	return nil
}

func (wl webLoader) loadPageContentFromApi(title string) (body []byte, err error) {
	url := wl.pageUrl + title
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
