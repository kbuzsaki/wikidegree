/*
Implements a priority queue for managing concurrent loading
in iddfs.go
*/
package iddfs

import (
	"container/heap"
	"fmt"
	api "github.com/kbuzsaki/wikidegree/api"
)

func DfsQueue(input <-chan api.TitlePath, requests <-chan chan<- api.TitlePath) {
	// internal channel variable that will be nil if the pqueue is empty
	// this lets us avoid taking requests when we're unable to fulfill them
	var internalRequests <-chan chan<- api.TitlePath
	pqueue := make(titlePathQueue, 0)


	for {
		select {
		// receive a TitlePath, add it to the queue
		case titlePath := <-input:
			fmt.Println("Pushing:", titlePath)
			pqueue.push(titlePath)
			if internalRequests == nil {
				internalRequests = requests
			}
		// receive a request for a TitlePath, pop it off the queue
		case request := <-internalRequests:
			titlePath := pqueue.pop()
			fmt.Println("Popping:", titlePath)
			request <- titlePath
			if len(pqueue) == 0 {
				internalRequests = nil
			}
		}
	}
}


func TestPQueue() {
	pqueue := make(titlePathQueue, 0)

	fmt.Println("start:", pqueue)

	patterns := []api.TitlePath{
		{"apple", "banana"},
		{"apple", "grape"},
		{"carrot", "grape", "pear", "bear"},
		{"carrot"},
		{"carrot", "grape", "pear"},
	}

	for i, pattern := range patterns {
		pqueue.push(pattern)
		fmt.Println("i:", i, "pattern:", pattern)
		fmt.Printf("%#v\n", pqueue)
	}

	item := pqueue.pop()
	fmt.Println(item)

	fmt.Println("pushing stuff")

	pqueue.push(api.TitlePath{"grape"})
	pqueue.push(api.TitlePath{"grape", "circus", "some", "long", "thing"})

	for len(pqueue) > 0 {
		item := pqueue.pop()
		fmt.Println(item)
	}
}


// Code implementing a priority queue, based off of
// https://golang.org/pkg/container/heap/#pkg-overview
//
// The type isn't exported because it's an implementation detail.
// Anyone wanting to use this should go through the
// DfsQueue function and communicate using channels
type titlePathQueue []api.TitlePath

// The methods that I actually care about for the DfsQueue
// They're preferable to heap.Interface's equivalents
// because they use the TitlePath type instead of interface{}
func (pq *titlePathQueue) push(item api.TitlePath) {
	heap.Push(pq, item)
}

func (pq *titlePathQueue) pop() api.TitlePath {
	item := heap.Pop(pq)
	return item.(api.TitlePath)
}

// The methods needed to implement heap.Interface
func (h titlePathQueue) Len() int {
	return len(h)
}

func (h titlePathQueue) Less(i, j int) bool {
	// heap implements a min-heap but we want longer TitlePaths
	// to be popped first, so use greater than instead of less than
	return len(h[i]) > len(h[j])
}

func (h titlePathQueue) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (pqueue *titlePathQueue) Push(item interface{}) {
	*pqueue = append(*pqueue, item.(api.TitlePath))
}

func (pqueue *titlePathQueue) Pop() interface{} {
	length := len(*pqueue)
	item := (*pqueue)[length - 1]
	*pqueue = (*pqueue)[: length - 1]
	return item
}

