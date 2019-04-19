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

	SecondInterval = time.Second
)

var (
	ErrLtMinDelay = errors.New("lt delay min")
)

type TimeWheel struct {
	counter    int64
	interval   time.Duration
	ticker     *time.Ticker
	slots      []*Timer
	timer      map[interface{}]int
	currentPos int
	slotNum    int
	started    bool

	ctx    context.Context
	cancel context.CancelFunc

	updateEventChannel chan *Event
}

func NewTimeWheel(interval time.Duration, slotNum int) (*TimeWheel, error) {
	if interval < SecondInterval || slotNum <= 0 {
		return nil, ErrLtMinDelay
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
	return tw, nil
}

func (tw *TimeWheel) initSlots() {
	for i := 0; i < tw.slotNum; i++ {
		tw.slots[i] = NewTimer()
	}
}

func (tw *TimeWheel) Start() {
	if tw.started {
		return
	}

	tw.currentPos = tw.getInitPosition()
	tw.ticker = time.NewTicker(tw.interval / 2)
	tw.started = true
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
		pos   = tw.getWritePosition(delay)
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
	var (
		pos   = tw.getWritePosition(delay)
		timer = tw.slots[pos]
	)

	timer.Sleep(delay)
}

func (tw *TimeWheel) After(delay time.Duration) <-chan time.Time {
	var (
		pos   = tw.getWritePosition(delay)
		timer = tw.slots[pos]
	)

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
			go tw.tickHandler()

		case <-tw.updateEventChannel:

		case <-tw.ctx.Done():
			return
		}
	}
}

func (tw *TimeWheel) tickHandler() {
	pos := tw.getInitPosition()
	tw.currentPos = pos
	timer := tw.slots[pos]
	timer.LoopOnce()

	// wheel full, reset init 0
	// if tw.currentPos == tw.slotNum-1 {
	// 	tw.currentPos = 0
	// } else {
	// 	tw.currentPos++
	// }
}

func (tw *TimeWheel) GetTimers() []*Timer {
	return tw.slots
}

type TimerStatsRes struct {
	SlotID int
	Len    int
}

func (tw *TimeWheel) GetTimersLength() []TimerStatsRes {
	var res = make([]TimerStatsRes, tw.slotNum)
	for idx, tm := range tw.slots {
		res[idx] = TimerStatsRes{
			SlotID: idx,
			Len:    tm.Len(),
		}
	}

	return res
}

func (tw *TimeWheel) getPosition(d time.Duration) (pos int) {
	return tw.callGetPosition(d, posReadMode)
}

func (tw *TimeWheel) getWritePosition(d time.Duration) (pos int) {
	return tw.callGetPosition(d, posWriteMode)
}

func (tw *TimeWheel) callGetPosition(delay time.Duration, mode int) int {
	var (
		pos       int
		plus      int
		delayUnit int
	)

	if tw.interval >= time.Millisecond && tw.interval < time.Second {
		delayUnit = int(delay.Nanoseconds() / 1000 / 1000)
		plus = int(time.Now().Unix()) + delayUnit
	} else {
		// defualt second unit
		delayUnit = int(delay.Seconds())
		plus = int(time.Now().Unix()) + delayUnit
	}

	pos = plus % tw.slotNum

	if mode == posWriteMode && pos == tw.currentPos {
		pos++
	}

	return pos
}

func (tw *TimeWheel) getInitPosition() int {
	var pos int
	if tw.interval >= time.Millisecond && tw.interval < time.Second {
		pos = int(time.Now().Nanosecond()/1000/1000) % tw.slotNum
	} else {
		pos = int(time.Now().Unix()) % tw.slotNum
	}

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
