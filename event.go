package timewheel

import (
	"fmt"
	"time"
)

// ExpireFunc represents a function will be executed when a event is trigged.
type ExpireFunc func()

// An Event represents an elemenet of the events in the timer.
type Event struct {
	slotPos int // mark timeWheel slot index
	index   int // index in the min heap structure

	ttl    time.Duration // wait delay time
	expire time.Time     // due timestamp
	fn     ExpireFunc    // callback function

	next    *Event
	cron    bool // repeat task
	cronNum int  // cron circle num
	alone   bool // indicates event is alone or in the free linked-list of timer
}

// Less is used to compare expiration with other events.
func (e *Event) Less(o *Event) bool {
	return e.expire.Before(o.expire)
}

// Delay is used to give the duration that event will expire.
func (e *Event) Delay() time.Duration {
	return e.expire.Sub(time.Now())
}

func (e *Event) String() string {
	return fmt.Sprintf("index %d ttl %v, expire at %v", e.index, e.ttl, e.expire)
}
