package processors

import (
	"github.com/kbuzsaki/wikidegree/batch"
	"github.com/kbuzsaki/wikidegree/batch/helpers"
	"github.com/kbuzsaki/wikidegree/wiki"
)

var present = []byte("a")

func NewBlobReverseLinker(config batch.Config, out chan<- wiki.Page) (batch.PageProcessor, error) {
	return &blobReverseLinker{config: config, out: out}, nil
}

func NewFilteringBlobReverseLinker(config batch.Config, predicate helpers.PagePredicate, out chan<- wiki.Page) (batch.PageProcessor, error) {
	return &blobReverseLinker{predicate: predicate, config: config, out: out}, nil
}

type blobReverseLinker struct {
	predicate helpers.PagePredicate
	config    batch.Config
	out       chan<- wiki.Page
}

func (cl *blobReverseLinker) Setup() error {
	return nil
}

func (cl *blobReverseLinker) ProcessPage(page wiki.Page) error {
	for _, link := range page.Links {
		linkedPage := wiki.Page{Title: link}

		if cl.predicate == nil || cl.predicate(linkedPage) {
			linkedPage.SetBlob(page.Title, present)
			cl.out <- linkedPage
		}
	}

	return nil
}

func (cl *blobReverseLinker) Teardown() error {
	close(cl.out)
	return nil
}
