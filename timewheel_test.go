package timewheel

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeWheelBase(t *testing.T) {
	tw, err := NewTimeWheel(time.Second, 60)
	if err != nil {
		t.Error(err)
	}
	tw.Start()

	var b = false
	tw.AfterFunc(1*time.Second, func() {
		b = true
	})

	time.Sleep(2 * time.Second)
	assert.True(t, b)
}

func TestTimeWheelRemove(t *testing.T) {
	tw, err := NewTimeWheel(time.Second, 60)
	if err != nil {
		t.Error(err)
	}
	tw.Start()

	var b = false
	entry, _ := tw.AfterFunc(1*time.Second, func() {
		b = true
	})

	tw.RemoveTimer(entry.event)
	time.Sleep(2 * time.Second)
	assert.False(t, b)
}

func TestTimeWheelAfter(t *testing.T) {
	tw, err := NewTimeWheel(time.Second, 60)
	if err != nil {
		t.Error(err)
	}
	tw.Start()

	// var b = false
	select {
	case <-tw.After(1 * time.Second):
	case <-time.After(2 * time.Second):
		t.Error("after func timeout")
	}
}
