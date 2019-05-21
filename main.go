package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

type (
	item     string // food item
	channels struct {
		burgers chan<- item
		fries   chan<- item
		sodas   chan<- item
	}
)

var wgConsumers sync.WaitGroup
var wgChefAlarm sync.WaitGroup

var done = make(chan bool)
var msgs = make(chan int)

func minMapValue(m map[item]int) int {
	min := math.MaxInt32

	for _, v := range m {
		if v < min {
			min = v
		}
	}

	return min
}

func mapToString(m map[item]int) string {
	s := ""
	for k, v := range m {
		s += fmt.Sprintf(" %s=%d", k, v)
	}
	return s
}

func chef(c int, ch channels) {
	defer close(ch.burgers)
	defer close(ch.fries)
	defer close(ch.sodas)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// only produce burgers, fries, sodas
	burger := item("burger")
	fries := item("fries")
	soda := item("soda")

	for i := 0; i < c; i++ {
		// produce random combination of two non-same items
		switch r.Intn(3) {
		case 0:
			wgChefAlarm.Add(2)
			ch.burgers <- burger
			ch.fries <- fries
		case 1:
			wgChefAlarm.Add(2)
			ch.burgers <- burger
			ch.sodas <- soda
		case 2:
			wgChefAlarm.Add(2)
			ch.sodas <- soda
			ch.fries <- fries
		}
		wgChefAlarm.Wait() //  wait for items to be consumed
	}
}

func consumer(name string, channelA <-chan item, channelB <-chan item) {
	defer wgConsumers.Done()

	counts := map[item]int{}

	for {
		select {
		case item := <-channelA:
			if item == "" { // channel is closed
				channelA = nil
			} else {
				wgChefAlarm.Done()
				counts[item]++
			}
		case item := <-channelB:
			if item == "" { // channel is closed
				channelB = nil
			} else {
				wgChefAlarm.Done()
				counts[item]++
			}
		}
		if (channelA == nil) && (channelB == nil) {
			// all channels closed
			fmt.Printf(
				"%s ate %d times (item counts%s)\n",
				name,
				minMapValue(counts),
				mapToString(counts),
			)
			break
		}
	}
}

func main() {

	burgers := make(chan item)
	fries := make(chan item)
	sodas := make(chan item)

	ch := channels{
		burgers: burgers,
		fries:   fries,
		sodas:   sodas,
	}

	go chef(100, ch)

	wgConsumers.Add(3)
	go consumer("Johnny Burger Goode", fries, sodas)
	go consumer("Francis R. Yes     ", burgers, sodas)
	go consumer("S. Oda             ", burgers, fries)

	wgConsumers.Wait()
}
