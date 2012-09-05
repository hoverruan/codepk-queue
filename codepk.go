package main

import (
	"fmt"
	"sync"
	"time"
)

const COUNT = 100000000
const MAX_WIN = 2
const READ_THREADS = 2

var failed int = 0
var now int64
var wg sync.WaitGroup

func pop(q chan int64) {
	for ts := range q {
		wg.Done()
		if (now - ts) > MAX_WIN {
			failed++
		}
	}
}

func main() {
	q := make(chan int64)
	now = time.Now().Unix()
	wg.Add(COUNT)

	ticket := time.Tick(1 * time.Second)
	go func() {
		for t := range ticket {
			now = t.Unix()
		}
	}()

	for i := 0; i < READ_THREADS; i++ {
		go pop(q)
	}

	for i := 0; i < COUNT; i++ {
		q <- now
	}
	close(q)

	wg.Wait()
	fmt.Println("Done.", "Failed:", failed)
}
