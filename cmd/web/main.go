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

const (
	timeLimit = time.Second * 10
)

func lookup(writer http.ResponseWriter, request *http.Request) {
	values := request.URL.Query()
	start := values.Get("start")
	end := values.Get("end")

	startTime := time.Now()
	path, err := lookupPathWithTimeout(start, end)
	duration := time.Since(startTime)

	result := make(map[string]interface{})
	result["time"] = duration.String()
	if err != nil {
		log.Print(err)
		result["error"] = err.Error()
	} else {
		result["path"] = path
	}

	resultBytes, _ := json.Marshal(&result)
	io.WriteString(writer, string(resultBytes))
}

func lookupPathWithTimeout(start, end string) (wiki.TitlePath, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeLimit)
	defer cancel()

	result, err := lookupPath(ctx, start, end)

	if ctx.Err() != nil {
		log.Println("timed out :(")
		return nil, errors.New("Timed out after 10 seconds.")
	} else if err != nil {
		log.Println("Got error:", err)
		return nil, err
	} else {
		log.Println("Got result:", result)
		return result, nil
	}
}

func lookupPath(ctx context.Context, start, end string) (wiki.TitlePath, error) {
	// valiate start and end titles exist
	if start == "" || end == "" {
		return nil, errors.New("start and end parameters required")
	}
	start = wiki.EncodeTitle(start)
	end = wiki.EncodeTitle(end)

	// load the page loader, currently only bolt
	pageLoader, err := wiki.GetBoltPageLoader()
	if err != nil {
		return nil, err
	}
	defer pageLoader.Close()

	// validate that the start page exists and has links
	startPage, err := pageLoader.LoadPage(start)
	if err != nil {
		return nil, err
	}
	if len(startPage.Links) == 0 {
		return nil, errors.New("start page has no links!")
	}

	// validate that the end page exists
	endPage, err := pageLoader.LoadPage(end)
	if err != nil {
		return nil, err
	}

	// use the page titles instead of the user input in case there were redirects
	start = wiki.EncodeTitle(startPage.Title)
	end = wiki.EncodeTitle(endPage.Title)

	log.Println("Finding path from '" + start + "' to '" + end + "'")

	// actually find the path using bfs
	pathFinder := bfs.GetBfsPathFinder(pageLoader)
	return pathFinder.FindPath(ctx, start, end)
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/api/lookup", lookup)
	http.ListenAndServe(":8080", nil)
}
