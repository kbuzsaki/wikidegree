package processors

import (
	"strings"

	"github.com/kbuzsaki/wikidegree/batch"
	"github.com/kbuzsaki/wikidegree/wiki"
)

func NewDeadLinkFilterer(config batch.Config, pr wiki.PageRepository, out chan<- wiki.Page) (batch.PageProcessor, error) {
	return &deadLinkFilter{pr: pr, config: config, out: out}, nil
}

type deadLinkFilter struct {
	pr     wiki.PageRepository
	config batch.Config
	out    chan<- wiki.Page
}

func (cl *deadLinkFilter) Setup() error {
	return nil
}

func (cl *deadLinkFilter) Process(page wiki.Page) error {
	validLinks, updated, err := cl.getValidLinks(page)
	if err != nil {
		return err
	}

	if updated {
		page.Links = validLinks
		cl.out <- page
	}

	return nil
}

func (cl *deadLinkFilter) Teardown() error {
	close(cl.out)
	return nil
}

func (cl *deadLinkFilter) getValidLinks(page wiki.Page) ([]string, bool, error) {
	var validLinks []string

	updated := false
	for _, link := range page.Links {
		_, err := cl.pr.LoadPage(link)

		if err != nil && !strings.HasPrefix(err.Error(), "No entry") {
			return nil, false, err
		} else if err != nil {
			if cl.config.Debug {
				//log.Printf("found dead link %#v in page %#v\n", link, page.Title)
			}
			updated = true
		} else {
			validLinks = append(validLinks, link)
		}
	}

	return validLinks, updated, nil
}
