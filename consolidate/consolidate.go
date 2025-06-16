// This package provides consolidation for two types of events:
//   - events that are function calls (use Func constructor for such a cases)
//   - events which are chan messages (use Chan constructor for such a cases)
//
// In both cases the events comes from origin source will be consolidated and passed to destination source wia following procedure:
//   - when some series of origin events are separated by long intervals without events the each such bunch of events will be translated to single consolidated event
//   - if the sequence of origin events has no intervals without events then the consolidated event will be generated with some frequency that is less than frequency of origin events.
//
// Both constructors has common parameters: delay and maxDelay (time.Duration values).
// delay - determines interval between last origin event into the bunch and consolidated event. Each new origin event resets the waiting for consolidated event to delay.
// The events that comes very frequently (with intervals between events less than delay) will newer generate the consolidated event by this mechanism.
// To solve this issue the another timer is set on the very first event after start or after last consolidated event.
// This timer is set on maxDelay and if the first mechanism fails the consolidated event will be generated when maxDelay expired.
// The only resalable relation between delay and maxDelay: delay must be less than maxDelay.
// So, the minimum interval between consolidated events is delay, while maxDelay determines the minimal interval between consolidated events in case of very frequent origin events.
//
// For case when origin events appears in bunches i.e. with short intervals between some amount of events and with big intervals between bunches
// the delay have to be slightly bigger that biggest short interval and maxDelay have to be bigger than duration of the longest events' bunch.
//
// When a flow of origin events has unpredictable intensity it worth to tune delay and maxDelay values via set of experiments with that flow.
package consolidate

import (
	"time"
)

type consolidate struct {
	inCh     chan struct{} // incoming event channel
	doFunc   func()        // function that is called on consolidated event
	cancel   chan struct{} // chan witch closure will stop the internal loop
	delay    time.Duration // bunch events consolidation interval
	maxDelay time.Duration // frequent events consolidation interval
}

// Func constructor creates consolidation function that have to be called when origin event appear. The function cFunc that will to be called on consolidated event is passed to the constructor.
func Func(delay, maxDelay time.Duration, cFunc func()) (func(), func()) {
	c := newConsolidator(delay, maxDelay, cFunc)
	return c.eventFunc, c.close
}

// Chan constructor creates consolidation chan that have to be filled by messages as origin events without blocking like:
//
//	select{
//	case ch <- struct{}{}:
//	default:
//	}
//
// The outCh that will be filled with consolidated events is passed to the constructor.
// outCh need to be buffered or received event must be handled quicker than delay (the minimal interval between consolidated events).
func Chan(delay, maxDelay time.Duration, outCh chan struct{}) (chan struct{}, func()) {
	c := newConsolidator(delay, maxDelay, outEventFunc(outCh))
	return c.inCh, c.close
}

// newConsolidator creates internal consolidate struct and starts the consolidation loop in separate goroutine.
func newConsolidator(delay, maxDelay time.Duration, do func()) *consolidate {
	c := &consolidate{
		delay:    delay,
		maxDelay: maxDelay,
		inCh:     make(chan struct{}, 1), // it have to be buffered to grantee that very first event is always stored
		doFunc:   do,
		cancel:   make(chan struct{}),
	}
	go c.loop()
	return c
}

// eventFunc is a functional wrapper for incoming events.
func (c *consolidate) eventFunc() {
	select {
	case c.inCh <- struct{}{}:
	default:
	}
}

// outEventFunc returns the function which call will send the event message to outCh without blocking.
func outEventFunc(outCh chan struct{}) func() {
	return func() {
		select {
		case outCh <- struct{}{}:
		default:
		}
	}
}

// close stops the internal loop to release all consolidate resources.
func (r *consolidate) close() {
	select {
	case <-r.cancel:
		return
	default:
		close(r.cancel)
	}
}

// loop - internal loop that provides event consolidation
func (c *consolidate) loop() {
	delayTimer := time.NewTimer(time.Hour)
	delayTimer.Stop()
	wdTimer := time.NewTimer(time.Hour)
	wdTimer.Stop()
	var wd <-chan time.Time
	for {
		select {
		case <-c.cancel:
			return
		case <-c.inCh:
			delayTimer.Reset(c.delay)
			if wd == nil { // it is first event after start or after last consolidated event
				wdTimer.Reset(c.maxDelay)
				wd = wdTimer.C
			}
		case <-delayTimer.C:
			c.doFunc() // consolidated event in bunch events mode
			wd = nil
		case <-wd:
			c.doFunc() // consolidated event in frequent events mode
			wd = nil
		}
	}
}
