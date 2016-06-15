package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"time"
	"net/http"
	bfs "github.com/kbuzsaki/wikidegree/bfs"
	api "github.com/kbuzsaki/wikidegree/api"
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

func lookupPathWithTimeout(start, end string) (api.TitlePath, error) {
	resultChan := make(chan api.TitlePath)
	errorChan := make(chan error)

	go func() {
		result, err := lookupPath(start, end)
		if err != nil {
			errorChan <- err
		}
		resultChan <- result
	}()

	timeout := time.After(10 * time.Second)
	select {
		case result := <-resultChan:
			log.Println("Got result:", result)
			return result, nil
		case err := <-errorChan:
			log.Println("Got error:", err)
			return nil, err
		case <-timeout:
			log.Println("timed out :(")
			return nil, errors.New("timed out after 10 seconds")
	}

	return nil, nil
}

func lookupPath(start, end string) (api.TitlePath, error) {
	// valiate start and end titles exist
	if start == "" || end == "" {
		return nil, errors.New("start and end parameters required")
	}
	start = api.EncodeTitle(start)
	end = api.EncodeTitle(end)

	// load the page loader, currently only bolt
	pageLoader, err := api.GetBoltPageLoader()
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
	start = api.EncodeTitle(startPage.Title)
	end = api.EncodeTitle(endPage.Title)

	log.Println("Finding path from '" + start + "' to '" + end + "'")

	// actually find the path using bfs
	pathFinder := bfs.GetBfsPathFinder(pageLoader)
	return pathFinder.FindPath(start, end)
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/api/lookup", lookup)
	http.ListenAndServe(":8000", nil)
}
