package main

import (
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/gosuri/uilive"
	"golang.org/x/exp/rand"
)

type Desc struct {
	Probability float64 // [0,1]
	Rate        float64 // [0,1]
}

type Person struct {
	Capability float64 // [0,1]
	Satisfy    Desc
	Disappoint Desc
}

// there is probability of p to get 1, otherwise 0.
func Random(p float64) float64 {
	if p > rand.Float64() {
		return 1
	}
	return 0
}

func S(x float64) float64 {
	if x >= 0 {
		return 1
	}
	return -1
}

func min(vs ...float64) float64 {
	m := vs[0]
	for _, v := range vs {
		if v < m {
			m = v
		}
	}
	return m
}

func f(caps, expect_prob_rate float64) (happy_rate float64) {
	const N = 1000000
	E := 0.1
	B := caps

	count := 0
	if B-E >= 0 {
		count++
	}
	for i := 2; i <= N; i++ {
		E = E + S(B-E)*min(E, 1-E)*expect_prob_rate

		if B-E >= 0 {
			count++
		}
	}

	return float64(count) / float64(N)
}

type Item struct {
	Caps      float64
	EP        float64
	HappyRate float64
}

func main() {
	const N = 30
	var arr []float64
	for i := 0; i <= N; i++ {
		arr = append(arr, float64(i)*float64(1)/float64(N))
	}

	generateCh := make(chan Item, 100)
	happyRateCh := make(chan Item, 100)

	SIZE := len(arr) * len(arr)
	go func(out chan<- Item) {
		defer close(out)
		for _, caps := range arr {
			for _, expect_prob_rate := range arr {
				out <- Item{
					Caps: caps,
					EP:   expect_prob_rate,
				}
			}
		}
	}(generateCh)

	go func(gonum int) {
		defer close(happyRateCh)

		var wg sync.WaitGroup
		for i := 0; i < gonum; i++ {
			wg.Add(1)
			go func(in <-chan Item, out chan<- Item) {
				defer wg.Done()

				for item := range in {
					item.HappyRate = f(item.Caps, item.EP)
					out <- item
				}
			}(generateCh, happyRateCh)
		}

		wg.Wait()
	}(8)

	uilive.Out = os.Stderr
	writer := uilive.New()
	writer.Start()

	items := make([]Item, 0, SIZE)
	count := 0
	for item := range happyRateCh {
		count++
		items = append(items, item)

		rate := float64(count) / float64(SIZE)

		if count%128 == 0 {
			fmt.Fprintf(writer, "Generate.. (%d/%d) %f%%\n", count, SIZE, rate*100)
		}
	}
	fmt.Fprintln(writer, "Done!")

	sort.Slice(items, func(i, j int) bool {
		return items[i].HappyRate < items[j].HappyRate
	})

	for _, item := range items {
		caps := item.Caps
		expect_prob_rate := item.EP
		happy_rate := item.HappyRate
		fmt.Printf("%f\t%f\t%f\n", caps, expect_prob_rate, happy_rate)
	}
}
