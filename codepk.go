package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	COUNT         = 10000000
	READ_THREADS  = 10
	WRITE_THREADS = 10
)

type IQueue interface {
	push(int64)
	pop() int64
	stop()
}

// internal channel queue

type ChannelQueue struct {
	queue chan int64
}

func NewChannelQueue() *ChannelQueue {
	q := new(ChannelQueue)
	q.queue = make(chan int64)

	return q
}

func (self *ChannelQueue) push(val int64) {
	self.queue <- val
}

func (self *ChannelQueue) pop() int64 {
	return <-self.queue
}

func (self *ChannelQueue) stop() {
	close(self.queue)
}

func runTest(q IQueue) {
	now := time.Now().Unix()

	ticket := time.Tick(1 * time.Second)
	go func() {
		for t := range ticket {
			now = t.Unix()
		}
	}()

	for i := 0; i < READ_THREADS; i++ {
		go func() {
			for {
				ts := q.pop()
				if ts <= 0 {
					break
				}
			}
		}()
	}

	writeWg := new(sync.WaitGroup)
	writeWg.Add(WRITE_THREADS)
	for i := 0; i < WRITE_THREADS; i++ {
		go func() {
			for j := 0; j < COUNT/WRITE_THREADS; j++ {
				q.push(now)
			}
			writeWg.Done()
		}()
	}
	writeWg.Wait()
	q.stop()

	fmt.Println("Done.")
}

func main() {
	q := NewChannelQueue()

	runTest(q)
}
