package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/kbuzsaki/wikidegree/wiki"
)

const defaultXmlDumpFilename = "xml/enwiki-20151201-pages-articles.xml"

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

	go loadPagesFromXml(xmlDumpFilename, pages, redirects)

	pageSaver, err := wiki.GetBoltPageSaver(indexFilename, redirFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer pageSaver.Close()

	counter := 0
	for {
		select {
		case page := <-pages:
			err = pageSaver.SavePage(page)
			if err != nil {
				log.Fatal(err)
			}
		case redirect := <-redirects:
			err = pageSaver.SaveRedirect(redirect.Title, redirect.Redir.Title)
			if err != nil {
				log.Fatal(err)
			}
		}

		counter++
		if counter%1000 == 0 {
			fmt.Println(counter)
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

func loadPagesFromXml(filename string, pages chan wiki.Page, redirects chan Page) {
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
}
