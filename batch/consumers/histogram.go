package consumers

import (
	"fmt"
	"sort"
	"sync"

	"github.com/kbuzsaki/wikidegree/batch"
)

func HistogramInts(wg *sync.WaitGroup, config batch.Config, ints <-chan int) {
	defer wg.Done()

	histogram := make(map[int]int)

	for i := range ints {
		if count, ok := histogram[i]; ok {
			histogram[i] = count + 1
		} else {
			histogram[i] = 1
		}
	}

	keys := make([]int, 0, len(histogram))
	for key := range histogram {
		keys = append(keys, key)
	}
	sort.Ints(keys)

	fmt.Println("histogram:")
	pages := 0
	sum := 0
	for _, key := range keys {
		val := histogram[key]
		pages += val
		sum += key * val

		fmt.Printf("%v: %v\n", key, val)
	}
	fmt.Printf("pages: %v\n", pages)
	fmt.Printf("sum: %v\n", sum)
}
