package processors

import (
	"log"
	"strings"

	"github.com/kbuzsaki/wikidegree/batch"
	"github.com/kbuzsaki/wikidegree/wiki"
)

func NewDeadTitleFilterer(config batch.Config, pr wiki.PageRepository, out chan<- string) (batch.TitleProcessor, error) {
	return &deadTitleFilter{pr: pr, config: config, out: out}, nil
}

type deadTitleFilter struct {
	pr     wiki.PageRepository
	config batch.Config
	out    chan<- string
}

func (cl *deadTitleFilter) Setup() error {
	return nil
}

func (cl *deadTitleFilter) ProcessTitle(title string) error {
	_, err := cl.pr.LoadPage(title)
	if err != nil && strings.HasPrefix(err.Error(), "error loading") {
		if cl.config.Debug {
			log.Printf("Found dead title: %#v, error: %s", title, err)
		}

		cl.out <- title
	} else if err != nil {
		return err
	}

	return nil
}

func (cl *deadTitleFilter) Teardown() error {
	close(cl.out)
	return nil
}
