package timewheel

import (
	"fmt"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func BenchmarkAddTimer(t *testing.B) {
	var timer = NewTimer()
	for index := 1; index < t.N; index++ {
		timer.Add(time.Millisecond*time.Duration(index), func() {
		})
	}
}
func BenchmarkAddDelTimer(t *testing.B) {
	var timer = NewTimer()
	for index := 1; index < t.N; index++ {
		ev := timer.Add(time.Millisecond*time.Duration(index), func() {
		})
		timer.Del(ev)
	}
}

func TestTimerLoop(t *testing.T) {
	timer := NewTimer()
	var wg sync.WaitGroup

	wg.Add(1)
	begin := time.Now()
	timer.Add(time.Millisecond*20, func() {
		defer wg.Done()
		if elapsed := time.Since(begin); elapsed > 26*time.Millisecond {
			assert.Fail(t, fmt.Sprintf("expected execute event after 20 milliseconds, but actual after %v", elapsed.String()))
		}
	})

	wg.Add(1)
	timer.Add(time.Millisecond*10, func() {
		defer wg.Done()
		if elapsed := time.Since(begin); elapsed > 16*time.Millisecond {
			assert.Fail(t, fmt.Sprintf("expected execute event after 10 milliseconds, but actual after %v", elapsed.String()))
		}
	})

	event := timer.Events()[1]
	timer.Start()

	// event expired in loop
	wg.Wait()
	assert.Equal(t, 0, timer.Len())

	// event recyling
	assert.NotEqual(t, unsafe.Pointer(timer.free), unsafe.Pointer(event))
	timer.Del(event)
	assert.Equal(t, unsafe.Pointer(timer.free), unsafe.Pointer(event))
	assert.Nil(t, timer.free.fn)

	// reset recyled event
	e1 := timer.Add(time.Millisecond*20, func() {
		t.Fatal("run")
	})
	e2 := timer.Add(time.Millisecond*10, func() {
		t.Fatal("run")
	})
	timer.Del(e1)
	timer.Del(e2)
	time.Sleep(time.Millisecond * 26)

	// update event
	u1b := false
	u1 := timer.Add(time.Millisecond*20, func() {
		u1b = true
	})
	timer.Set(u1, time.Millisecond*5)
	time.Sleep(time.Millisecond * 10)
	assert.True(t, u1b)
}
