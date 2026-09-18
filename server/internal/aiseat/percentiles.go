package aiseat

import (
	"sort"
	"time"
)

// percentiles.go is #505 part 1: the runner's own decision-latency
// distribution, readable while it plays.
//
// The number that matters at a table is not the mean. A bot whose
// mean decision is 40ms and whose p999 is nine seconds is a bot that
// visibly hangs twice a game, and the mean says nothing about it.
// ADR 0033 §10's whole posture — "the table never waits on a bot" —
// is a claim about the tail, so the tail is what gets measured.

// latencyRing is how many recent decisions a Runner keeps. 1024 is
// about six turns of a four-seat table: enough for a stable p999,
// small enough (16 KiB) to carry on every runner without thinking
// about it.
const latencyRing = 1024

// Percentiles is a latency distribution by nearest rank.
//
// Count is how many samples the distribution was computed from, NOT
// how many decisions the runner has made: the ring holds the most
// recent latencyRing of them and a longer game rolls off the front.
// A reader comparing Count with Stats.Decisions is comparing a window
// with a total, which is why both are reported.
type Percentiles struct {
	Count int           `json:"count"`
	P50   time.Duration `json:"p50"`
	P99   time.Duration `json:"p99"`
	P999  time.Duration `json:"p999"`
	Max   time.Duration `json:"max"`
}

// PercentilesOf computes the distribution of d by nearest rank
// (the smallest sample at or above the pth percentile), on a COPY:
// the caller's slice is not reordered.
//
// Nearest rank rather than interpolation because these are latencies
// and every reported value should be a latency that actually
// happened. An interpolated p999 of 4.7s on a sample where the two
// worst decisions took 2s and 9s describes neither.
func PercentilesOf(d []time.Duration) Percentiles {
	if len(d) == 0 {
		return Percentiles{}
	}
	s := append([]time.Duration(nil), d...)
	sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })
	return Percentiles{
		Count: len(s),
		P50:   nearestRank(s, 0.50),
		P99:   nearestRank(s, 0.99),
		P999:  nearestRank(s, 0.999),
		Max:   s[len(s)-1],
	}
}

// nearestRank returns the ceil(p·N)-th smallest sample of a sorted
// slice, 1-indexed and clamped. s must be non-empty.
func nearestRank(s []time.Duration, p float64) time.Duration {
	n := len(s)
	rank := int(float64(n)*p + 0.9999999)
	if rank < 1 {
		rank = 1
	}
	if rank > n {
		rank = n
	}
	return s[rank-1]
}

// observeLatency records one whole-decision duration in the ring.
func (r *Runner) observeLatency(d time.Duration) {
	r.latMu.Lock()
	defer r.latMu.Unlock()
	if len(r.latencies) < latencyRing {
		r.latencies = append(r.latencies, d)
		return
	}
	r.latencies[r.latNext] = d
	r.latNext = (r.latNext + 1) % latencyRing
}

// latencySnapshot copies the ring.
func (r *Runner) latencySnapshot() []time.Duration {
	r.latMu.Lock()
	defer r.latMu.Unlock()
	return append([]time.Duration(nil), r.latencies...)
}
