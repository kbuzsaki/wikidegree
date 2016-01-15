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
	"regexp"
	"strings"
	"net/url"
)

type TitlePath []string

func (titlePath TitlePath) Head() string {
	return titlePath[len(titlePath) - 1]
}

func (titlePath TitlePath) Catted(title string) TitlePath {
	newTitlePath := make(TitlePath, len(titlePath), len(titlePath) + 1)
	copy(newTitlePath, titlePath)
	newTitlePath = append(newTitlePath, title)
	return newTitlePath
}

type Page struct {
	Title   string
	Content string
}

type ParsedPage struct {
	Title string
	Links []string
}


func ParsePage(page Page) ParsedPage {
	regex, _ := regexp.Compile("\\[\\[(.+?)(\\]\\]|\\||#)")

	matches := regex.FindAllStringSubmatch(page.Content, -1)

	var links []string
	for _, match := range matches {
		link := match[1]
		link = EncodeTitle(link)
		links = append(links, link)
	}

	return ParsedPage{page.Title, links}
}

func EncodeTitle(title string) string {
	// the first character of the string is case insensitive,
	// but all the rest is *sensitive*
	title = strings.ToUpper(title[0:1]) + title[1:]
	title = strings.Replace(title, " ", "_", -1)
	title = url.QueryEscape(title)
	return title
}

