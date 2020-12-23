package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type SlidingWindow struct {
	buffer           []bool
	len              int
	curr             int
	faultCnt         int
	fuseRate         float64
	state            int
	lastFuseOpenTime time.Time
	recoveryTime     int64
	lock             sync.Mutex
}

const DEFAULT_BUFFER_SIZE = 11
const DEFAULT_FUSE_RATE = 0.8
const DEFAULT_RECOVERY_TIME = 10000
const (
	SLIDING_WINDOW_FUSE_STATE_CLOSED   = iota
	SLIDING_WINDOW_FUSE_STATE_OPEN     = iota
	SLIDING_WINDOW_FUSE_STATE_HALFOPEN = iota
)

func NewSlidingWindow() *SlidingWindow {
	ref := &SlidingWindow{}
	len := DEFAULT_BUFFER_SIZE
	ref.len = len
	ref.fuseRate = DEFAULT_FUSE_RATE
	ref.buffer = make([]bool, len)
	ref.recoveryTime = DEFAULT_RECOVERY_TIME

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond).C
		for {
			select {
			case <-ticker:
				if ref.state == SLIDING_WINDOW_FUSE_STATE_OPEN {
					now := time.Now()
					fmt.Println(now.Sub(ref.lastFuseOpenTime).Milliseconds())
					if now.Sub(ref.lastFuseOpenTime).Milliseconds() > ref.recoveryTime {
						ref.clearBuffer()
						ref.state = SLIDING_WINDOW_FUSE_STATE_CLOSED
					}
				}
			}
		}
	}()
	return ref
}

func (ref *SlidingWindow) calc() bool {
	return float64(ref.faultCnt)/float64(ref.len) > ref.fuseRate
}

func (ref *SlidingWindow) clearBuffer() {
	for i := 0; i < ref.len; i++ {
		ref.buffer[i] = false
	}
	ref.curr = 0
}

func (ref *SlidingWindow) Put(val bool) bool {
	if ref.state == SLIDING_WINDOW_FUSE_STATE_OPEN {
		return true
	}
	ref.lock.Lock()
	defer ref.lock.Unlock()
	if val {
		ref.faultCnt++
	}
	if ref.buffer[ref.curr] {
		ref.faultCnt--
	}
	ref.buffer[ref.curr] = val
	ref.curr = (ref.curr + 1) % ref.len
	if ref.calc() {
		fmt.Println("change state")
		ref.state = SLIDING_WINDOW_FUSE_STATE_OPEN
		ref.lastFuseOpenTime = time.Now()
		ref.faultCnt = 0
		ref.clearBuffer()
		return true
	}
	return false
}

func (ref *SlidingWindow) print() {
	for i := 0; i < ref.len; i++ {
		val := 0
		if ref.buffer[i] {
			val = 1
		}
		fmt.Printf("%v ", val)
	}
	fmt.Printf("\n")
}

func main() {
	slidingWindow := NewSlidingWindow()
	for ; ; {
		val := true
		if rand.Intn(100) < 90 {
			val = false
		}
		fmt.Println(slidingWindow.Put(val))
		time.Sleep(10 * time.Millisecond)
		slidingWindow.print()
	}
}
