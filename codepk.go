package main

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const (
	COUNT         = 1000000000
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

// ring array queue

const (
	RING_SIZE uint64 = 1024
	RING_MASK uint64 = RING_SIZE - 1
)

var ringReadBuffSize uint64 = uint64(runtime.NumCPU())

type RingArrayQueue struct {
	lastCommittedIndex uint64
	nextFreeIndex      uint64
	readerIndex        uint64
	contents           [RING_SIZE]int64
}

func NewRingArrayQueue() *RingArrayQueue {
	return &RingArrayQueue{lastCommittedIndex: 0, nextFreeIndex: 1, readerIndex: 1}
}

func (self *RingArrayQueue) push(val int64) {
	myIndex := atomic.AddUint64(&self.nextFreeIndex, 1) - 1

	for myIndex > (self.readerIndex + RING_SIZE - 256) {
		runtime.Gosched()
	}

	self.contents[myIndex&RING_MASK] = val

	for !atomic.CompareAndSwapUint64(&self.lastCommittedIndex, myIndex-1, myIndex) {
		runtime.Gosched()
	}
}

func (self *RingArrayQueue) pop() int64 {
	myIndex := atomic.AddUint64(&self.readerIndex, 1) - 1

	for myIndex > (self.lastCommittedIndex + ringReadBuffSize) {
		runtime.Gosched()
	}

	return self.contents[myIndex&RING_MASK]
}

func (self *RingArrayQueue) stop() {
}

// test runner

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
}

func main() {
	fmt.Println("Count:", COUNT, "Read threads", READ_THREADS, "Write threads", WRITE_THREADS)

	t0 := time.Now()
	runTest(NewChannelQueue())

	t1 := time.Now()
	fmt.Printf("Channel queue: %v\n", t1.Sub(t0))

	runTest(NewRingArrayQueue())

	t2 := time.Now()
	fmt.Printf("Ring array queue: %v\n", t2.Sub(t1))
}
