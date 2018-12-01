package main

import (
	"fmt"
	"math"
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

func f(caps, expect_prob_rate float64) (happy_rate float64) {
	const N = 100000
	E := 0.1
	B := caps

	count := 0
	if B-E >= 0 {
		count++
	}
	for i := 2; i <= N; i++ {
		//m := math.Max(B, 1-B)
		//	E = E + ((math.Abs(B-E)+1*m)/(2*m))*S(B-E)*math.Min(E, 1-E)*expect_prob_rate
		E = E + S(B-E)*math.Min(math.Min(E, 1-E), 0.01)*expect_prob_rate
		//E = E + ((math.Abs(B-E)+0.5)/(math.Abs(B-E)+1))*S(B-E)*math.Min(math.Min(E, 1-E), 0.1)*expect_prob_rate
		//Mood := B - E

		if B-E >= 0 {
			count++
		}
		//fmt.Println(i, E, Mood, count, i, float64(count)*100/float64(i), "%")
	}

	return float64(count) / float64(N)
}

type Item struct {
	Caps      float64
	EP        float64
	HappyRate float64
}

func main() {
	const N = 100
	var arr []float64
	for i := 0; i <= N; i++ {
		arr = append(arr, float64(i)*float64(1)/float64(N))
	}
	//fmt.Println(arr)
	//os.Exit(0)

	generateCh := make(chan Item, 10000)
	happyRateCh := make(chan Item, 10000)

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
	// start listening for updates and render
	writer.Start()

	items := make([]Item, 0, SIZE)
	count := 0
	for item := range happyRateCh {
		count++
		items = append(items, item)

		rate := float64(count) / float64(SIZE)

		if count%500 == 0 {
			fmt.Fprintf(writer, "Generate.. (%d/%d) %f%%\n", count, SIZE, rate*100)
		}
	}
	fmt.Fprintln(writer, "Done!")

	sort.Slice(items, func(i, j int) bool {
		return items[i].HappyRate < items[j].HappyRate
	})

	//fmt.Println("caps\tprob\trate\texpect_prob_rate\thappy_rate")
	for _, item := range items {
		caps := item.Caps
		expect_prob_rate := item.EP
		happy_rate := item.HappyRate
		//fmt.Println("caps:", caps, "prob;", prob, "rate:", rate, "expect_prob_rate:", prob*rate, "\t", happy_rate )
		//	fmt.Printf("%f\t%f\t%f\t%f\t%f\n", caps, prob, rate, prob*rate, happy_rate)
		fmt.Printf("%f\t%f\t%f\n", caps, expect_prob_rate, happy_rate)
	}
	/*
		prob_rates := []float64{0, 0.5, 1}
		//caps := []float64{0.1, 0.3, 0.5, 0.7, 0.9}
		caps := []float64{0.5, 0.7, 0.9}
		N := 100 // event number

		personsCh := make(chan Person)

		go func(persons chan<- Person) {
			defer close(persons)
			for _, dp_prob := range prob_rates {
				for _, dp_rate := range prob_rates {
					for _, st_prob := range prob_rates {
						for _, st_rate := range prob_rates {
							for _, capability := range caps {
								persons <- Person{
									Disappoint: Desc{
										Probability: dp_prob,
										Rate:        dp_rate,
									},
									Satisfy: Desc{
										Probability: st_prob,
										Rate:        st_rate,
									},
									Capability: capability,
								}
							}
						}
					}
				}
			}
		}(personsCh)

		var persons []Person
		for person := range personsCh {
			persons = append(persons, person)
		}

		const InitExpect = 0.2

		size := len(persons) * len(persons)

		countCh := make(chan int)
		defer close(countCh)

		go func(countCh <-chan int) {
			count := 0
			ticker := time.NewTicker(time.Millisecond * 100)
			for {
				select {
				case c := <-countCh:
					count += c
				case <-ticker.C:
					fmt.Fprintln(os.Stderr, count, "/", size, "\t", float64(count)/float64(size))
				}
			}
		}(countCh)

		for _, p1 := range persons {
			for _, p2 := range persons {
				countCh <- 1
				fmt.Println(p1, p2)
				E_1 := InitExpect
				//E_2 := InitExpect
				B_1 := p2.Capability
				//	B_2 := p1.Capability

				for i := 0; i < N; i++ {
					E_1 = E_1 + epi(E_1-B_1)*E_1*p1.Satisfy.Rate*Random(p1.Satisfy.Probability) + epi_reverse(E_1-B_1)*E_1*p1.Disappoint.Rate*Random(p1.Disappoint.Probability)
					fmt.Println(i+1, E_1)
				}
			}
		}
	*/
}

func epi(x float64) float64 {
	if x >= 0 {
		return 1
	}
	return 0
}
func epi_reverse(x float64) float64 {
	switch epi(x) {
	case 1.0:
		return 0
	case 0.0:
		return 1
	}
	panic("you shouldn't be here!")
}
