package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kbuzsaki/wikidegree/wiki"
)

const defaultXmlDumpFilename = "xml/enwiki-20151201-pages-articles.xml"
const printThresh = 100
const bufferMax = 1000

func main() {
	xmlDumpFilename := flag.String("xml", defaultXmlDumpFilename, "the full text xml dump to import from")
	indexFilename := flag.String("index", wiki.DefaultIndexName, "the boltdb index db")
	redirFilename := flag.String("redir", wiki.DefaultRedirName, "the boltdb redirect db")

	flag.Parse()

	fmt.Println("Starting...")
	load(*xmlDumpFilename, *indexFilename, *redirFilename)
}

func load(xmlDumpFilename, indexFilename, redirFilename string) {
	pages := make(chan wiki.Page)
	redirects := make(chan Page)
	done := make(chan struct{})

	go loadPagesFromXml(xmlDumpFilename, pages, redirects, done)

	pageSaver, err := wiki.GetBoltPageSaver(indexFilename, redirFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer pageSaver.Close()

	var pageBuffer []wiki.Page
	var redirectBuffer []wiki.Redirect

	counter := 0
	start := time.Now()
	for {
		select {
		case page := <-pages:
			pageBuffer = append(pageBuffer, page)

			if len(pageBuffer) >= bufferMax {
				err = pageSaver.SavePages(pageBuffer)
				if err != nil {
					log.Fatal(err)
				}
				pageBuffer = pageBuffer[0:0]
			}
		case redirect := <-redirects:
			redirectBuffer = append(redirectBuffer, wiki.Redirect{redirect.Title, redirect.Redir.Title})

			if len(redirectBuffer) >= bufferMax {
				err = pageSaver.SaveRedirects(redirectBuffer)
				if err != nil {
					log.Fatal(err)
				}
				redirectBuffer = redirectBuffer[0:0]
			}
		case <-done:
			err = pageSaver.SavePages(pageBuffer)
			if err != nil {
				log.Fatal(err)
			}
			err = pageSaver.SaveRedirects(redirectBuffer)
			if err != nil {
				log.Fatal(err)
			}
			break
		}

		counter++
		if counter%printThresh == 0 {
			duration := time.Since(start)
			fmt.Println(counter, "(", duration, ")")
			start = time.Now()
		}
	}
}

type Redirect struct {
	Title string `xml:"title,attr"`
}

type Page struct {
	Title string   `xml:"title"`
	Redir Redirect `xml:"redirect"`
	Text  string   `xml:"revision>text"`
}

func loadPagesFromXml(filename string, pages chan wiki.Page, redirects chan Page, done chan struct{}) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	decoder := xml.NewDecoder(file)

	var nilredir Redirect
	for {
		token, _ := decoder.Token()
		if token == nil {
			break
		}

		switch element := token.(type) {
		case xml.StartElement:
			if element.Name.Local == "page" {
				var page Page
				decoder.DecodeElement(&page, &element)

				if page.Redir != nilredir {
					redirects <- page
				} else {
					links := wiki.ParseLinks(page.Text)
					pages <- wiki.Page{page.Title, page.Title, links}
				}
			}
		}
	}

	done <- struct{}{}
}
