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
	redirFilename := flag.String("redir", wiki.DefaultRedirName, "the boltdb redirect db")
	flag.Parse()

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	fmt.Println("Starting...")
	load(*xmlDumpFilename, *indexFilename, *redirFilename)
}

func load(xmlDumpFilename, indexFilename, redirFilename string) {
	xmlPages := make(chan XmlPage, 1000)
	pages := make(chan []wiki.Page, 1000)
	redirects := make(chan []wiki.Redirect, 1000)
	done := make(chan struct{})

	wg := &sync.WaitGroup{}
	wg.Add(3)
	go loadPagesFromXml(wg, xmlDumpFilename, xmlPages)
	go aggregatePages(wg, xmlPages, pages, redirects, done)
	go savePages(wg, indexFilename, redirFilename, pages, redirects, done)
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

func aggregatePages(wg *sync.WaitGroup, xmlPages <-chan XmlPage, pages chan<- []wiki.Page, redirects chan<- []wiki.Redirect, done chan<- struct{}) {
	defer wg.Done()

	var pageBuffer []wiki.Page
	var redirectBuffer []wiki.Redirect

	counter := 0
	start := time.Now()

	var nilredir XmlRedirect
	for xmlPage := range xmlPages {
		if xmlPage.Redirect != nilredir {
			redirect := wiki.Redirect{xmlPage.Title, xmlPage.Redirect.Title}
			redirectBuffer = append(redirectBuffer, redirect)

			if len(redirectBuffer) >= bufferMax {
				redirects <- redirectBuffer
				redirectBuffer = nil
			}
		} else {
			links := wiki.ParseLinks(xmlPage.Text)
			page := wiki.Page{xmlPage.Title, xmlPage.Title, links}
			pageBuffer = append(pageBuffer, page)

			if len(pageBuffer) >= bufferMax {
				pages <- pageBuffer
				pageBuffer = nil
			}
		}

		counter++

		if counter%printThresh == 0 {
			fmt.Println(counter, "(", time.Since(start), ")")
			start = time.Now()
		}
	}

	redirects <- redirectBuffer
	close(redirects)
	pages <- pageBuffer
	close(pages)
	done <- struct{}{}
}

func savePages(wg *sync.WaitGroup, indexFilename, redirFilename string, pages <-chan []wiki.Page, redirects <-chan []wiki.Redirect, done <-chan struct{}) {
	defer wg.Done()

	pageSaver, err := wiki.GetBoltPageSaver(indexFilename, redirFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer pageSaver.Close()

	for {
		select {
		case pageBuffer := <-pages:
			err := pageSaver.SavePages(pageBuffer)
			if err != nil {
				log.Fatal(err)
			}
		case redirectBuffer := <-redirects:
			err := pageSaver.SaveRedirects(redirectBuffer)
			if err != nil {
				log.Fatal(err)
			}
		case <-done:
			break
		}
	}
}
