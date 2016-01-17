package bfs

import (
	"fmt"
	api "github.com/kbuzsaki/wikidegree/api"
)

// queue for use by serial bfs
type TitlePathQueue []api.TitlePath

func (pathQueue *TitlePathQueue) Push(titlePath api.TitlePath) {
	*pathQueue = append(*pathQueue, titlePath)
}

func (pathQueue *TitlePathQueue) Pop() api.TitlePath {
	var titlePath api.TitlePath
	titlePath, *pathQueue = (*pathQueue)[0], (*pathQueue)[1:]
	return titlePath
}

// serial implementation of bfs
func (bpf *bfsPathFinder) findNearestPathSerial(start string, end string) api.TitlePath {
	visited := make(map[string]bool)
	visited[start] = true
	frontier := TitlePathQueue{{start}}

	for len(frontier) > 0 {
		titlePath := frontier.Pop()

		fmt.Println("Loading:", titlePath)
		if page, err := bpf.pageLoader.LoadPage(titlePath.Head()); err == nil {
			for _, title := range page.Links {
				newTitlePath := titlePath.Catted(title)

				if title == end {
					return newTitlePath
				} else if !visited[title] {
					visited[title] = true
					frontier.Push(newTitlePath)
				}
			}
		} else {
			fmt.Println("Failed to load: ", titlePath.Head())
		}
	}

	return nil
}
