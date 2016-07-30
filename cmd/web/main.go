package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/kbuzsaki/wikidegree/wiki"
	"github.com/kbuzsaki/wikidegree/search/bfs"
	"golang.org/x/net/context"
)

type Server struct {
	pageLoader wiki.PageLoader
}

func (s *Server) renderJSON(writer http.ResponseWriter, resp interface{}) {
	respBytes, _ := json.Marshal(&resp)
	io.WriteString(writer, string(respBytes))
}

func (s *Server) renderError(writer http.ResponseWriter, err error) {
	s.renderJSON(writer, map[string]string{"error": err.Error()})
}

func (s *Server) handleLookup(writer http.ResponseWriter, request *http.Request) {
	values := request.URL.Query()
	start := values.Get("start")
	end := values.Get("end")

	startTime := time.Now()
	path, err := s.lookupPathWithTimeout(start, end, 10 * time.Second)
	duration := time.Since(startTime)

	if err != nil {
		s.renderError(writer, err)
	} else {
		s.renderJSON(writer, map[string]interface{}{
			"time": duration.String(),
			"path": path,
		})
	}
}

func (s *Server) lookupPathWithTimeout(start, end string, timeout time.Duration) (wiki.TitlePath, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	result, err := s.lookupPath(ctx, start, end)
	if ctx.Err() != nil {
		return nil, errors.New("Timed out after 10 seconds.")
	}

	return result, err
}

func (s *Server) lookupPath(ctx context.Context, start, end string) (wiki.TitlePath, error) {
	// valiate start and end titles exist
	if start == "" || end == "" {
		return nil, errors.New("start and end parameters required")
	}
	start = wiki.NormalizeTitle(start)
	end = wiki.NormalizeTitle(end)

	// validate that the start page exists and has links
	startPage, err := s.pageLoader.LoadPage(start)
	if err != nil {
		return nil, err
	}
	if len(startPage.Links) == 0 {
		return nil, errors.New("start page has no links!")
	}

	// validate that the end page exists
	endPage, err := s.pageLoader.LoadPage(end)
	if err != nil {
		return nil, err
	}

	// use the page titles instead of the user input in case there were redirects
	start = wiki.NormalizeTitle(startPage.Title)
	end = wiki.NormalizeTitle(endPage.Title)

	log.Println("Finding path from '" + start + "' to '" + end + "'")

	// actually find the path using bfs
	pathFinder := bfs.GetBfsPathFinder(s.pageLoader)
	return pathFinder.FindPath(ctx, start, end)
}

func (s *Server) handleLinks(writer http.ResponseWriter, request *http.Request) {
	values := request.URL.Query()
	title := values.Get("title")
	if title == "" {
		s.renderError(writer, errors.New("title is required"))
		return
	}

	title = wiki.EncodeTitle(title)
	page, err := s.pageLoader.LoadPage(title)
	if err != nil {
		s.renderError(writer, err)
	} else {
		s.renderJSON(writer, page)
	}
}

func main() {
	pageLoader, err := wiki.GetBoltPageLoader()
	if err != nil {
		log.Fatal(err)
	}

	s := &Server{pageLoader}

	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/api/lookup", s.handleLookup)
	http.HandleFunc("/api/page", s.handleLinks)
	http.ListenAndServe(":8080", nil)
}
