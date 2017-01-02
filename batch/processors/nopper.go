// titleNopper and pageNopper are the minimal possible TitleProcessor and PageProcessor
// they do nothing at all with their input and demonstrate the upper bound for batch throughput

package processors

import (
	"github.com/kbuzsaki/wikidegree/batch"
	"github.com/kbuzsaki/wikidegree/wiki"
)

func NewTitleNopper(config batch.Config) (batch.TitleProcessor, error) {
	return &titleNopper{config: config}, nil
}

type titleNopper struct {
	config batch.Config
}

func (cl *titleNopper) Setup() error {
	return nil
}

func (cl *titleNopper) ProcessTitle(title string) error {
	return nil
}

func (cl *titleNopper) Teardown() error {
	return nil
}

func NewPageNopper(config batch.Config) (batch.PageProcessor, error) {
	return &pageNopper{config: config}, nil
}

type pageNopper struct {
	config batch.Config
}

func (cl *pageNopper) Setup() error {
	return nil
}

func (cl *pageNopper) ProcessPage(page wiki.Page) error {
	return nil
}

func (cl *pageNopper) Teardown() error {
	return nil
}
