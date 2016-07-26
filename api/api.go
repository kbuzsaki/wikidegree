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
	"net/url"
	"regexp"
	"strings"
)

// Represents a wiki page
// Contains the page's unique title and the titles of all of the pages that it
// links to.
type Page struct {
	Redirector string   // the original link used to get to the page, usually but not always the same as title
	Title      string   // the actual title of the page
	Links      []string // the links on the page
}

// Represents something that can load wiki pages
// Takes the title of the page and returns the Page struct.
type PageLoader interface {
	LoadPage(title string) (Page, error)
	Close()
}

// Represents a series of page titles/links that take you from one page
// to another.
type TitlePath []string

func (titlePath TitlePath) Head() string {
	return titlePath[len(titlePath)-1]
}

func (titlePath TitlePath) Catted(title string) TitlePath {
	newTitlePath := make(TitlePath, len(titlePath), len(titlePath)+1)
	copy(newTitlePath, titlePath)
	newTitlePath = append(newTitlePath, title)
	return newTitlePath
}

// Represents something that, given a PageLoader, can look up a path from one
// page to another
type PathFinder interface {
	SetPageLoader(pageLoader PageLoader)
	FindPath(start, end string) (TitlePath, error)
}

// Helper function that parses the links from a page's body text.
func ParseLinks(content string) []string {
	regex, _ := regexp.Compile("\\[\\[(.+?)(\\]\\]|\\||#)")

	matches := regex.FindAllStringSubmatch(content, -1)

	var links []string
	for _, match := range matches {
		link := match[1]
		link = EncodeTitle(link)
		links = append(links, link)
	}

	return links
}

// Helper function that formats and encodes a page title for web lookup
func EncodeTitle(title string) string {
	// the first character of the string is case insensitive,
	// but all the rest is *sensitive*
	title = strings.ToUpper(title[0:1]) + title[1:]
	title = strings.Replace(title, " ", "_", -1)
	title = url.QueryEscape(title)
	return title
}
