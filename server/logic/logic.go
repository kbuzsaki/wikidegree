package logic

import (
	"context"
	"errors"
	"log"

	"github.com/kbuzsaki/wikidegree/search/bfs"
	"github.com/kbuzsaki/wikidegree/wiki"
)

type Logic interface {
	LookupPath(ctx context.Context, start, end string) (wiki.TitlePath, error)
	LookupPage(ctx context.Context, title string) (wiki.Page, error)
}

type logicImpl struct {
	pageLoader wiki.PageLoader
	pathFinder wiki.PathFinder
}

func New() (Logic, error) {
	pageLoader, err := wiki.GetBoltPageLoader()
	if err != nil {
		return nil, err
	}
	pathFinder := bfs.GetBfsPathFinder(pageLoader)

	return &logicImpl{pageLoader, pathFinder}, nil
}

func (l *logicImpl) LookupPath(ctx context.Context, start, end string) (wiki.TitlePath, error) {
	startPage, err := l.LookupPage(ctx, start)
	if err != nil {
		return nil, err
	}
	if len(startPage.Links) == 0 {
		return nil, errors.New("start page has no links!")
	}

	endPage, err := l.LookupPage(ctx, end)
	if err != nil {
		return nil, err
	}

	// use the page titles instead of the user input in case there were redirects
	log.Println("Finding path from '" + startPage.Title + "' to '" + endPage.Title + "'")
	return l.pathFinder.FindPath(ctx, startPage.Title, endPage.Title)
}

func (l *logicImpl) LookupPage(ctx context.Context, title string) (wiki.Page, error) {
	if title == "" {
		return wiki.Page{}, errors.New("title required")
	}
	title = wiki.NormalizeTitle(title)

	return l.pageLoader.LoadPage(title)
}
