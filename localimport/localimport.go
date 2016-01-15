package main

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kbuzsaki/wikidegree/api"
)

const xmlDumpFilename = "xml/enwiki-20151201-pages-articles.xml"

const indexName = "db/index.db"
const redirName = "db/redir.db"

const commitThreshold = 10000

func main() {
	load()
}

func load() {
	fmt.Println("Starting...")

	pages := make(chan api.Page)
	redirects := make(chan Page)

	//go loadPagesFromMysql("kbuzsaki@/wiki", pages)
	go loadPagesFromXml(xmlDumpFilename, pages, redirects)

	index, err := bolt.Open(indexName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer index.Close()

	redir, err := bolt.Open(redirName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer redir.Close()

	counter := 0
	for {
		select {
		case page := <-pages:
			index.Update(func(tx *bolt.Tx) error {
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
		case redirect := <-redirects:
			redir.Update(func(tx *bolt.Tx) error {
				titleBytes := []byte(redirect.Title)
				bucket, err := tx.CreateBucket(titleBytes)
				if err != nil {
					log.Fatal(err)
				}

				linkBytes := []byte(redirect.Redir.Title)
				bucket.Put(linkBytes, []byte{})
				return nil
			})
		}

		counter++
		fmt.Println(counter)
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

func loadPagesFromXml(filename string, pages chan api.Page, redirects chan Page) {
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
					links := api.ParseLinks(page.Text)
					pages <- api.Page{page.Title, links}
				}
			}
		}
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

		links := api.ParseLinks(body)
		page := api.Page{title, links}
		pages <- page
	}
}
