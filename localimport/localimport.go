package main

import (
	"database/sql"
	"fmt"
	"log"
	"github.com/boltdb/bolt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kbuzsaki/wikidegree/api"
	"os"
	"encoding/xml"
)

const xmlDumpFilename = "enwiki-20151201-pages-articles.xml"

func main() {
	fmt.Println("Starting...")

	pages := make(chan api.Page)
	parsed := make(chan api.ParsedPage)
	//go loadPagesFromMysql("kbuzsaki@/wiki", pages)
	go loadPagesFromXml(xmlDumpFilename, pages)
	go parsePages(pages, parsed)

	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	for page := range parsed {
		db.Update(func(tx *bolt.Tx) error {
			titleBytes := []byte(page.Title)
			bucket, err := tx.CreateBucket(titleBytes)
			if err != nil {
				log.Fatal(err)
			}

			for _, link := range page.Links {
				linkBytes := []byte(link)
				bucket.Put(linkBytes, []byte{})
			}
			return nil
		})
	}
}

type Redirect struct {
    Title string `xml:"title,attr"`
}

type Page struct {
    Title string `xml:"title"`
    Redir Redirect `xml:"redirect"`
    Text string `xml:"revision>text"`
}

func loadPagesFromXml(filename string, pages chan api.Page) {
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
				var page Page
				decoder.DecodeElement(&page, &element)

				page.Text = "[Text]"
				fmt.Println(page)
				var nilredir Redirect
				fmt.Println(page.Redir == nilredir)
			}
		}
	}
}

func parsePages(pages chan api.Page, parsed chan api.ParsedPage) {
	for page := range pages {
		parsed <- api.ParsePage(page)
	}
}

func loadPagesFromMysql(dataSource string, pages chan api.Page) {
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

		page := api.Page{title, body}
		pages <- page
	}
}

