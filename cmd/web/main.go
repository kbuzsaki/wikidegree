package main

import (
	"net/http"

	"log"

	"github.com/kbuzsaki/wikidegree/server"
)

func main() {
	s, err := server.New()
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/api/path", s.HandlePathLookup)
	http.HandleFunc("/api/page", s.HandlePageLookup)
	http.ListenAndServe(":8080", nil)
}
