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

type testCase struct {
	name             string
	refreshSeq       int
	seqInterval      time.Duration
	seqCount         int
	expectedMaxDelay time.Duration
	expectedCount    int
}

var tCases = []testCase{
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
	assert.Eventually(t, receive(res), 30*time.Millisecond, 2*time.Millisecond)
	for range 2 {
		assert.Eventually(t, receive(res), 130*time.Millisecond, 5*time.Millisecond)
	}
	assert.Never(t, receive(res), 120*time.Millisecond, 5*time.Millisecond)
}

func testSequence(t *testing.T, result chan struct{}, tc testCase) {
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
}

func TestNewConsolidator(t *testing.T) {
	result := make(chan struct{}, 100)
	cFunc := func() { result <- struct{}{} }
	c := newConsolidator(minInterval, maxInterval, cFunc)
	defer c.close()
	for _, tc := range tCases {
		t.Run(tc.name, func(t *testing.T) {
			defer eventGenerator(tc.seqInterval, tc.refreshSeq, tc.seqCount, c.eventFunc)()
			testSequence(t, result, tc)
		})
	}
	t.Run("multi_close", func(t *testing.T) {
		assert.NotPanics(t, c.close)
		assert.NotPanics(t, c.close)
	})
}

func TestConsolidateFunc(t *testing.T) {
	result := make(chan struct{}, 100)
	cFunc := func() { result <- struct{}{} }
	origFunc, cancel := Func(minInterval, maxInterval, cFunc)
	defer cancel()
	for _, tc := range tCases {
		t.Run(tc.name, func(t *testing.T) {
			defer eventGenerator(tc.seqInterval, tc.refreshSeq, tc.seqCount, origFunc)()
			testSequence(t, result, tc)
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
			testSequence(t, result, tc)
		})
	}
}
