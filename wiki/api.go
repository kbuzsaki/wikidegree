/*
Implements a simple interface for loading the content from wikipedia pages.
This lets you easily find all of the pages that a given wikipedia page
links to.

Currently loads data a bit messily using the json wiki.

In the future, an alternate implementation could load the data from a local
copy of wikipedia, even extending the interface to allow searching for all
links to a particular page.
*/
package wiki

import (
	"context"
	"io"
	"regexp"
	"strings"
)

// Represents a wiki page
// Contains the page's unique title and the titles of all of the pages that it
// links to.
type Page struct {
	Redirector string   // the original link used to get to the page, usually but not always the same as title
	Title      string   // the actual title of the page
	Redirect   string   // the page that this page redirects to
	Links      []string // pages that this page links to
	Linkers    []string // pages that link to this page
}

// Represents something that can load wiki pages
// Takes the title of the page and returns the Page struct.
type PageLoader interface {
	LoadPage(title string) (Page, error)
	LoadPages(titles []string) ([]Page, error)
	io.Closer
}

type PageSaver interface {
	SavePage(page Page) error
	SavePages(pages []Page) error
	io.Closer
}

type PagePager interface {
	FirstPage() (Page, error)
	NextPage(title string) (Page, error)
	NextPages(title string, count int) ([]Page, error)
	io.Closer
}

// TODO: make this just the union of the above interfaces once https://github.com/golang/go/issues/6977 is fixed
type PageRepository interface {
	LoadPage(title string) (Page, error)
	LoadPages(titles []string) ([]Page, error)
	SavePage(page Page) error
	SavePages(pages []Page) error
	FirstPage() (Page, error)
	NextPage(title string) (Page, error)
	NextPages(title string, count int) ([]Page, error)
	NextTitles(title string, count int) ([]string, error)
	io.Closer
}

// Represents a series of page titles/links that take you from one page
// to another.
type TitlePath []string

func (titlePath TitlePath) Head() string {
	if len(titlePath) == 0 {
		return ""
	}
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
	FindPath(ctx context.Context, start, end string) (TitlePath, error)
}

// Helper function that parses the links from a page's body text.
func ParseLinks(content string) []string {
	if content == "" {
		return []string{}
	}

	regex, _ := regexp.Compile(`\[\[(.+?)(\]\]|\||#)`)

	matches := regex.FindAllStringSubmatch(content, -1)

	var links []string
	for _, match := range matches {
		link := match[1]
		link = NormalizeTitle(link)
		links = append(links, link)
	}

	return links
}

func NormalizeTitle(title string) string {
	if title == "" {
		return ""
	}

	title = strings.ToUpper(title[0:1]) + title[1:]
	title = strings.Replace(title, " ", "_", -1)

	return title
}
