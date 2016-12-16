package main

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/kbuzsaki/wikidegree/wiki"
)

const xmlDumpFilename = "xml/enwiki-20151201-pages-articles.xml"

func main() {
	load()
}

func load() {
	fmt.Println("Starting...")

	pages := make(chan wiki.Page)
	redirects := make(chan Page)

	//go loadPagesFromMysql("kbuzsaki@/wiki", pages)
	go loadPagesFromXml(xmlDumpFilename, pages, redirects)

	pageSaver, err := wiki.GetBoltPageSaver()
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

func loadPagesFromMysql(dataSource string, pages chan wiki.Page) {
	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT title, body FROM Pages")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var title string
		var body string
		if err := rows.Scan(&title, &body); err != nil {
			log.Fatal(err)
		}

		links := wiki.ParseLinks(body)
		page := wiki.Page{title, title, links}
		pages <- page
	}
}
