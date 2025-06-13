package consolidate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	minInterval = 100 * time.Millisecond
	maxInterval = 5 * minInterval
)

var tCases = []struct {
	name             string
	refreshSeq       int
	seqInterval      time.Duration
	seqCount         int
	expectedMaxDelay time.Duration
	expectedCount    int
}{
	{
		name:             "low",
		refreshSeq:       5,
		seqInterval:      2 * minInterval,
		seqCount:         0,
		expectedMaxDelay: minInterval + 20*time.Millisecond,
		expectedCount:    0,
	},
	{
		name:             "high",
		refreshSeq:       2,
		seqInterval:      minInterval / 2,
		seqCount:         0,
		expectedMaxDelay: maxInterval + 20*time.Millisecond,
		expectedCount:    0,
	},
	{
		name:             "two_lov",
		refreshSeq:       3,
		seqInterval:      2 * minInterval,
		seqCount:         3,
		expectedMaxDelay: minInterval + 20*time.Millisecond,
		expectedCount:    3,
	},
	{
		name:             "two_high",
		refreshSeq:       3,
		seqInterval:      minInterval / 2,
		seqCount:         10,
		expectedMaxDelay: maxInterval + 20*time.Millisecond,
		expectedCount:    2,
	},
}

func eventGenerator(interval time.Duration, count, series int, event func()) func() {
	ticker := time.NewTicker(interval)
	sCnt := 0
	stop := make(chan struct{})
	go func() {
		for {
			for range count {
				event()
				time.Sleep(5 * time.Microsecond)
			}
			sCnt++
			if series > 0 && sCnt >= series {
				return
			}
			select {
			case <-ticker.C:
				continue
			case <-stop:
				return
			}
		}
	}()
	return func() {
		select {
		case <-stop:
		default:
			close(stop)
		}
	}
}

func receive(ch chan struct{}) func() bool {
	return func() bool {
		select {
		case <-ch:
			return true
		default:
			return false
		}
	}
}

func TestEventGenerator(t *testing.T) {
	res := make(chan struct{}, 50)
	defer eventGenerator(100*time.Millisecond, 1, 3, func() { res <- struct{}{} })()
	for range 3 {
		assert.Eventually(t, receive(res), 130*time.Millisecond, 5*time.Millisecond)
	}
	assert.Never(t, receive(res), 120*time.Millisecond, 5*time.Millisecond)
}

func TestConsolidateFunc(t *testing.T) {
	result := make(chan struct{}, 100)
	cFunc := func() { result <- struct{}{} }
	origFunc, cancel := Func(minInterval, maxInterval, cFunc)
	defer cancel()
	for _, tc := range tCases {
		t.Run(tc.name, func(t *testing.T) {
			defer eventGenerator(tc.seqInterval, tc.refreshSeq, tc.seqCount, origFunc)()
			start := time.Now()
			<-result
			elapsed := time.Since(start)
			assert.LessOrEqual(t, elapsed, tc.expectedMaxDelay)
			if tc.expectedCount > 0 {
				for i := range tc.expectedCount - 1 {
					assert.Eventually(t, receive(result), tc.seqInterval+tc.expectedMaxDelay, 5*time.Millisecond, i)
				}
				assert.Never(t, receive(result), tc.seqInterval+tc.expectedMaxDelay, 5*time.Millisecond)
			}
		})
	}
}

func TestConsolidateChan(t *testing.T) {
	result := make(chan struct{}, 100)
	origCh, cancel := Chan(minInterval, maxInterval, result)
	defer cancel()
	for _, tc := range tCases {
		t.Run(tc.name, func(t *testing.T) {
			defer eventGenerator(tc.seqInterval, tc.refreshSeq, tc.seqCount, func() {
				select {
				case origCh <- struct{}{}:
				default:
				}
			})()
			start := time.Now()
			<-result
			elapsed := time.Since(start)
			assert.LessOrEqual(t, elapsed, tc.expectedMaxDelay)
			if tc.expectedCount > 0 {
				for i := range tc.expectedCount - 1 {
					assert.Eventually(t, receive(result), tc.seqInterval+tc.expectedMaxDelay, 5*time.Millisecond, i)
				}
				assert.Never(t, receive(result), tc.seqInterval+tc.expectedMaxDelay, 5*time.Millisecond)
			}
		})
	}
}
