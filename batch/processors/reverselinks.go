package processors

import (
	"github.com/kbuzsaki/wikidegree/batch"
	"github.com/kbuzsaki/wikidegree/wiki"
)

var present = []byte("a")

func NewBlobReverseLinker(config batch.Config, pr wiki.PageRepository, out chan<- wiki.Page) (batch.PageProcessor, error) {
	return &blobReverseLinker{pr: pr, config: config, out: out}, nil
}

type blobReverseLinker struct {
	pr     wiki.PageRepository
	config batch.Config
	out    chan<- wiki.Page
}

func (cl *blobReverseLinker) Setup() error {
	return nil
}

func (cl *blobReverseLinker) ProcessPage(page wiki.Page) error {
	for _, link := range page.Links {
		linkedPage := wiki.Page{Title: link}
		linkedPage.SetBlob(page.Title, present)
		cl.out <- linkedPage
	}

	return nil
}

func (cl *blobReverseLinker) Teardown() error {
	close(cl.out)
	return nil
}


