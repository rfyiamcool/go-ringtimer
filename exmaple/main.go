package main

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/rfyiamcool/go-timewheel"
)

func base() {
	tw := timewheel.NewTimeWheel(time.Second, 60)
	tw.Start()

	var b = false
	tw.AfterFunc(1*time.Second, func() {
		b = true
	})

	// max delay + 2
	time.Sleep(3 * time.Second)
	fmt.Println(b)
}

func multi() {
	tw := timewheel.NewTimeWheel(time.Millisecond, 60)
	tw.Start()

	var incr int32 = 0

	go func() {
		start := time.Now()
		for index := 1; index <= 60; index++ {
			tw.AfterFunc(time.Duration(index)*time.Millisecond, func() {
				atomic.AddInt32(&incr, 1)
			})
		}
		fmt.Println("multi add time cost: ", time.Now().Sub(start))
	}()

	// max delay + 2
	time.Sleep(63 * time.Millisecond)
	fmt.Println(incr)
}

func batch() {
	tw := timewheel.NewTimeWheel(time.Millisecond, 60)
	tw.Start()

	var incr int32 = 0
	start := time.Now()
	for index := 1; index <= 60; index++ {
		for i := 0; i < 10000; i++ {
			tw.AfterFunc(time.Duration(index)*time.Millisecond, func() {
				atomic.AddInt32(&incr, 1)
			})
		}
	}
	fmt.Println("batch add time cost: ", time.Now().Sub(start))

	// max delay + 2
	time.Sleep(60 * 3 * time.Millisecond)
	fmt.Println(incr)
}

func main() {
	base()
	multi()
	batch()
}
