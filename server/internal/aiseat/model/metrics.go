package model

import (
	"sort"
	"sync"
	"time"
)

// metrics.go is the sub-PR's "per-decision instrumentation: layer
// used, latency, tokens, escalation reason".
//
// It is not optional decoration. The funnel is a COST argument — ADR
// 0033 §5 estimates "cents per game" and then says the estimate is
// "to be replaced with measurements from the first bot-vs-bot run" —
// and an argument nobody can check is a guess. Everything here exists
// so that after a game somebody can say which layer answered, how
// often the escalation triggers fired, what the calls cost and how
// long the table waited.

// Layer names, as they appear in a record and a log line.
const (
	LayerA = "A" // the rules filter answered
	LayerB = "B" // the heuristic answered
	LayerC = "C" // the model answered
)

// Fallback causes: why a window that reached Layer C came back with
// Layer B's answer anyway. Empty means Layer C's answer was used, or
// that the window never reached it.
const (
	// FallbackNoClient — the tier wants a model and none is
	// configured. This is a supported deployment, not a fault.
	FallbackNoClient = "no-client"
	// FallbackNoBudget — the runner's deadline left too little time
	// to try. No call was made.
	FallbackNoBudget = "no-budget"
	// FallbackError — the call failed: outage, HTTP error, timeout.
	FallbackError = "error"
	// FallbackMalformed — the reply was not an index.
	FallbackMalformed = "malformed"
	// FallbackOutOfRange — the reply was a number that is not a move.
	FallbackOutOfRange = "out-of-range"
	// FallbackPolicyError — Layer B itself failed, which is the one
	// case with nothing left underneath; the runner's own fallback
	// takes it from here.
	FallbackPolicyError = "policy-error"
)

// DecisionRecord is one window, after the fact.
type DecisionRecord struct {
	// Layer is which layer produced the answer.
	Layer string
	// Rule is the Layer A rule that fired, when Layer == LayerA.
	Rule string
	// Escalations are the triggers that fired, sorted. Empty means
	// the cheap model was the right one.
	Escalations []string
	// Model is the model id that was asked, empty when none was.
	Model string
	// Attempted is true when a model call was actually made — the
	// difference between "the model was wrong" and "the model was
	// never reached", which is the distinction the outage drill turns
	// on.
	Attempted bool
	// Fallback names why Layer C's answer was not used.
	Fallback string
	// TimedOut refines Fallback == FallbackError: the call did not
	// fail, it did not finish. On a self-hosted model this is the
	// failure mode to watch — every window timing out is a seat
	// playing Layer B under a model tier's name.
	TimedOut bool
	// Latency is the whole Decide call.
	Latency time.Duration
	// ModelLatency is the model call alone, zero when none was made.
	ModelLatency time.Duration
	// Usage is what the call cost.
	Usage Usage
	// Index and Reason are the decision that came out.
	Index  int
	Reason string
}

// Stats are the aggregate counters, readable while the game runs.
type Stats struct {
	// Windows is every Decide call.
	Windows int64
	// ByLayer counts windows by the layer that answered them.
	ByLayer map[string]int64
	// ByEscalation counts how many windows each trigger fired on. A
	// window with two triggers counts in both.
	ByEscalation map[string]int64
	// ByFallback counts the reasons Layer C's answer was discarded.
	ByFallback map[string]int64
	// Escalated counts windows where at least one trigger fired —
	// i.e. windows that asked the frontier model rather than the
	// cheap one. Counted once per window, unlike ByEscalation.
	Escalated int64
	// ModelCalls is how many calls were actually attempted.
	ModelCalls int64
	// ModelTimeouts is how many of them ran out of budget rather than
	// failing. Read it next to MaxModelLatency: a deployment whose
	// model cannot answer inside MaxThink shows up here as a number
	// climbing toward ModelCalls, and nowhere else.
	ModelTimeouts int64
	// Usage totals every call.
	Usage Usage
	// ModelLatency totals the time spent inside model calls, and
	// MaxModelLatency is the worst one — the number that says whether
	// MaxThink is in danger.
	ModelLatency    time.Duration
	MaxModelLatency time.Duration
}

// AbsorptionRate is the share of windows Layer A answered — ADR 0033
// §5's number, as this policy measured it.
func (s Stats) AbsorptionRate() float64 {
	if s.Windows == 0 {
		return 0
	}
	return float64(s.ByLayer[LayerA]) / float64(s.Windows)
}

// EscalationRate is the share of windows that survived Layer A and
// asked the frontier model rather than the cheap one. ADR 0033 §5
// estimates "~20% escalating" and says plainly that the estimate is
// to be replaced by a measurement; this is the measurement.
func (s Stats) EscalationRate() float64 {
	reached := s.ByLayer[LayerB] + s.ByLayer[LayerC]
	if reached == 0 {
		return 0
	}
	return float64(s.Escalated) / float64(reached)
}

// Triggers lists the escalation reasons seen, most frequent first.
func (s Stats) Triggers() []string {
	out := make([]string, 0, len(s.ByEscalation))
	for k := range s.ByEscalation {
		out = append(out, k)
	}
	sort.Slice(out, func(i, j int) bool {
		if s.ByEscalation[out[i]] != s.ByEscalation[out[j]] {
			return s.ByEscalation[out[i]] > s.ByEscalation[out[j]]
		}
		return out[i] < out[j]
	})
	return out
}

// recorder holds the counters and a bounded ring of recent records.
// One per policy, and a policy is one seat, so the lock is
// uncontended in practice; it is here because Stats is read from a
// test or an operator's goroutine while the seat plays.
type recorder struct {
	mu      sync.Mutex
	stats   Stats
	ring    []DecisionRecord
	keep    int
	nextIdx int
	full    bool
}

func newRecorder(keep int) *recorder {
	if keep <= 0 {
		keep = 256
	}
	return &recorder{
		keep: keep,
		ring: make([]DecisionRecord, keep),
		stats: Stats{
			ByLayer:      map[string]int64{},
			ByEscalation: map[string]int64{},
			ByFallback:   map[string]int64{},
		},
	}
}

func (r *recorder) record(rec DecisionRecord) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stats.Windows++
	r.stats.ByLayer[rec.Layer]++
	if len(rec.Escalations) > 0 {
		r.stats.Escalated++
	}
	for _, e := range rec.Escalations {
		r.stats.ByEscalation[e]++
	}
	if rec.Fallback != "" {
		r.stats.ByFallback[rec.Fallback]++
	}
	if rec.TimedOut {
		r.stats.ModelTimeouts++
	}
	if rec.Attempted {
		r.stats.ModelCalls++
		r.stats.ModelLatency += rec.ModelLatency
		if rec.ModelLatency > r.stats.MaxModelLatency {
			r.stats.MaxModelLatency = rec.ModelLatency
		}
	}
	r.stats.Usage.InputTokens += rec.Usage.InputTokens
	r.stats.Usage.OutputTokens += rec.Usage.OutputTokens
	r.stats.Usage.CacheReadTokens += rec.Usage.CacheReadTokens
	r.stats.Usage.CacheWriteTokens += rec.Usage.CacheWriteTokens
	r.stats.Usage.CachedPromptTokens += rec.Usage.CachedPromptTokens

	r.ring[r.nextIdx] = rec
	r.nextIdx = (r.nextIdx + 1) % r.keep
	if r.nextIdx == 0 {
		r.full = true
	}
}

func (r *recorder) snapshot() Stats {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := r.stats
	out.ByLayer = copyCounts(r.stats.ByLayer)
	out.ByEscalation = copyCounts(r.stats.ByEscalation)
	out.ByFallback = copyCounts(r.stats.ByFallback)
	return out
}

// records returns the retained records, oldest first.
func (r *recorder) records() []DecisionRecord {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.full {
		return append([]DecisionRecord(nil), r.ring[:r.nextIdx]...)
	}
	out := make([]DecisionRecord, 0, r.keep)
	out = append(out, r.ring[r.nextIdx:]...)
	out = append(out, r.ring[:r.nextIdx]...)
	return out
}

func copyCounts(m map[string]int64) map[string]int64 {
	out := make(map[string]int64, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
