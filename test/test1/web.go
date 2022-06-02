package main

import (
	"fmt"
	"sync"
)

type Sog struct {
	DOG string
	ver int
}
type Cat struct {
	Cat string
	ver int
}
type Bird struct {
	DOG string
	ver int
}

type Print struct {
	Po  string
	ver int
}

func Product(k chan string) {
	for {
		k <- "dog"
		k <- "cat"
		k <- "fish"
	}

}

func Consumer(k chan string) {
	for i := 0; i < len(k); i++ {
		fmt.Println(<-k)
	}
}

func main() {

	prochan := make(chan string, 3)
	wg := sync.WaitGroup{}
	for {
		wg.Add(2)
		go Product(prochan)
		go Consumer(prochan)
		wg.Done()
	}
}
