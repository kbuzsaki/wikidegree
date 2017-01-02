package processors

import (
	"github.com/kbuzsaki/wikidegree/batch"
	"github.com/kbuzsaki/wikidegree/wiki"
)

func NewLinkCounter(config batch.Config, out chan<- int) (batch.PageProcessor, error) {
	return &linkCounter{config: config, out: out}, nil
}

type linkCounter struct {
	config batch.Config
	out    chan<- int
}

func (cl *linkCounter) Setup() error {
	return nil
}

func (cl *linkCounter) ProcessPage(page wiki.Page) error {
	if !page.IsRedirect() && !page.IsRedirected() {
		cl.out <- len(page.Links)
	}

	return nil
}

func (cl *linkCounter) Teardown() error {
	close(cl.out)
	return nil
}
