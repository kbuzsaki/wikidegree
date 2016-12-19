package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sync"

	"time"

	"github.com/kbuzsaki/wikidegree/wiki"
)

const defaultXmlDumpFilename = "xml/enwiki-20151201-pages-articles.xml"
const printThresh = 10000
const bufferMax = 10000

func main() {
	xmlDumpFilename := flag.String("xml", defaultXmlDumpFilename, "the full text xml dump to import from")
	indexFilename := flag.String("index", wiki.DefaultIndexName, "the boltdb index db")
	flag.Parse()

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	fmt.Println("Starting...")
	load(*xmlDumpFilename, *indexFilename)
}

func load(xmlDumpFilename, indexFilename string) {
	xmlPages := make(chan XmlPage, 1000)
	pages := make(chan []wiki.Page, 1000)
	done := make(chan struct{})

	wg := &sync.WaitGroup{}
	wg.Add(3)
	go loadPagesFromXml(wg, xmlDumpFilename, xmlPages)
	go aggregatePages(wg, xmlPages, pages, done)
	go savePages(wg, indexFilename, pages, done)
	wg.Wait()
}

type XmlRedirect struct {
	Title string `xml:"title,attr"`
}

type XmlPage struct {
	Title    string      `xml:"title"`
	Redirect XmlRedirect `xml:"redirect"`
	Text     string      `xml:"revision>text"`
}

func loadPagesFromXml(wg *sync.WaitGroup, filename string, xmlPages chan<- XmlPage) {
	defer wg.Done()

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	decoder := xml.NewDecoder(file)

	for {
		token, _ := decoder.Token()
		if token == nil {
			break
		}

		switch element := token.(type) {
		case xml.StartElement:
			if element.Name.Local == "page" {
				var xmlPage XmlPage
				decoder.DecodeElement(&xmlPage, &element)
				xmlPages <- xmlPage
			}
		}
	}

	close(xmlPages)
}

func aggregatePages(wg *sync.WaitGroup, xmlPages <-chan XmlPage, pages chan<- []wiki.Page, done chan<- struct{}) {
	defer wg.Done()

	var pageBuffer []wiki.Page
	counter := 0
	start := time.Now()

	for xmlPage := range xmlPages {
		links := wiki.ParseLinks(xmlPage.Text)
		page := wiki.Page{Title: xmlPage.Title, Redirect: xmlPage.Redirect.Title, Links: links}
		pageBuffer = append(pageBuffer, page)

		if len(pageBuffer) >= bufferMax {
			pages <- pageBuffer
			pageBuffer = nil
		}

		counter++

		if counter%printThresh == 0 {
			fmt.Println(counter, "(", time.Since(start), ")")
			start = time.Now()
		}
	}

	pages <- pageBuffer
	done <- struct{}{}
}

func savePages(wg *sync.WaitGroup, indexFilename string, pages <-chan []wiki.Page, done <-chan struct{}) {
	defer wg.Done()

	pageSaver, err := wiki.GetBoltPageSaver(indexFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer pageSaver.Close()

	remaining := true
	for remaining {
		select {
		case pageBuffer := <-pages:
			err := pageSaver.SavePages(pageBuffer)
			if err != nil {
				log.Fatal(err)
			}
		case <-done:
			remaining = false
		}
	}
}
