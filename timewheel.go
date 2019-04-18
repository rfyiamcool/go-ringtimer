package timewheel

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

const (
	posWriteMode = iota
	posReadMode
)

var (
	ErrLtMinDelay = errors.New("delay min")
)

type TimeWheel struct {
	counter    int64
	interval   time.Duration
	ticker     *time.Ticker
	slots      []*Timer
	timer      map[interface{}]int
	currentPos int
	slotNum    int

	ctx    context.Context
	cancel context.CancelFunc

	updateEventChannel chan *Event
}

func NewTimeWheel(interval time.Duration, slotNum int) *TimeWheel {
	if interval <= 0 || slotNum <= 0 {
		return nil
	}

	var ctx, cancel = context.WithCancel(context.Background())
	tw := &TimeWheel{
		interval: interval,
		slots:    make([]*Timer, slotNum),
		timer:    make(map[interface{}]int),
		slotNum:  slotNum,

		ctx:    ctx,
		cancel: cancel,
	}

	tw.initSlots()
	return tw
}

func (tw *TimeWheel) initSlots() {
	for i := 0; i < tw.slotNum; i++ {
		tw.slots[i] = NewTimer()
	}
}

func (tw *TimeWheel) Start() {
	tw.currentPos = tw.getInitPosition()
	tw.ticker = time.NewTicker(tw.interval)
	go tw.start()
}

func (tw *TimeWheel) Stop() {
	tw.cancel()
}

func (tw *TimeWheel) ResetTimer(ev *Event, delay time.Duration) (*Event, bool) {
	if ev == nil {
		return nil, false
	}

	timer := tw.slots[ev.slotPos]
	timer.Del(ev)
	newEvent := timer.Add(delay, ev.fn)
	ev = nil
	return newEvent, true
}

func (tw *TimeWheel) AddTimer(delay time.Duration, fn ExpireFunc) (*Event, error) {
	if delay < time.Millisecond {
		return nil, ErrLtMinDelay
	}

	var (
		pos   = tw.getPosition(delay)
		timer = tw.slots[pos]
	)

	ev := timer.addAny(delay, fn, false)
	ev.slotPos = pos

	return ev, nil
}

func (tw *TimeWheel) RemoveTimer(ev *Event) {
	if ev == nil {
		return
	}

	timer := tw.slots[ev.slotPos]
	timer.Del(ev)
}

func (tw *TimeWheel) Sleep(delay time.Duration) {
	timer := tw.getTimerInSlot(delay)
	timer.Sleep(delay)
}

func (tw *TimeWheel) After(delay time.Duration) <-chan time.Time {
	timer := tw.getTimerInSlot(delay)
	return timer.After(delay)
}

func (tw *TimeWheel) AfterFunc(delay time.Duration, fn ExpireFunc) (*Event, error) {
	return tw.AddTimer(delay, fn)
}

func (tw *TimeWheel) GetTimerCount() int64 {
	return atomic.LoadInt64(&tw.counter)
}

func (tw *TimeWheel) start() {
	tw.tickHandler()
	for {
		select {
		case <-tw.ticker.C:
			tw.tickHandler()

		case <-tw.updateEventChannel:

		case <-tw.ctx.Done():
			return
		}
	}
}

func (tw *TimeWheel) tickHandler() {
	timer := tw.slots[tw.currentPos]
	timer.LoopOnce()
	if tw.currentPos == tw.slotNum-1 {
		tw.currentPos = 0
	} else {
		tw.currentPos++
	}
}

func (tw *TimeWheel) GetTimers(d time.Duration) []*Timer {
	return tw.slots
}

func (tw *TimeWheel) getPosition(d time.Duration) (pos int) {
	return tw.callGetPosition(d, posReadMode)
}

func (tw *TimeWheel) getWritePosition(d time.Duration) (pos int) {
	return tw.callGetPosition(d, posWriteMode)
}

func (tw *TimeWheel) callGetPosition(delay time.Duration, mode int) int {
	var (
		pos          int
		delaySeconds = int(delay.Seconds())
		plus         = int(time.Now().Unix()) + delaySeconds
	)

	pos = plus % tw.slotNum
	if mode == posWriteMode && pos == tw.currentPos {
		pos++
	}
	return pos
}

func (tw *TimeWheel) getInitPosition() int {
	var pos = int(time.Now().Unix()) % tw.slotNum
	return pos
}

func (tw *TimeWheel) getTimerInSlot(delay time.Duration) *Timer {
	pos := tw.getPosition(delay)
	return tw.slots[pos]
}

func (tw *TimeWheel) incr() {
	atomic.AddInt64(&tw.counter, 1)
}

func (tw *TimeWheel) deincr() {
	atomic.AddInt64(&tw.counter, -1)
}

func (tw *TimeWheel) loadIncr() int64 {
	return atomic.LoadInt64(&tw.counter)
}
