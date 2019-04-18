package main

import (
	"fmt"
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

	time.Sleep(2 * time.Second)
	fmt.Println(b)
}

func main() {
	base()
}
