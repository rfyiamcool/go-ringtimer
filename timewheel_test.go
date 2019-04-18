package timewheel

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeWheelBase(t *testing.T) {
	tw := NewTimeWheel(time.Second, 60)
	tw.Start()

	var b = false
	tw.AfterFunc(1*time.Second, func() {
		b = true
	})

	time.Sleep(3 * time.Second)
	assert.True(t, b)
}
