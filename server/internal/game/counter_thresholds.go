package game

import (
	"strconv"
	"sync"
)

// counter_thresholds.go — when a static ability's "as long as this has N
// or more counters" effect BEGAN (#3862, ADR 0067 amendment 2026-10-15).
//
//	CR 613.7  "An effect with an earlier timestamp is applied before an
//	           effect with a later timestamp."
//	CR 613.6  An effect that starts to apply in one layer carries on in
//	           the later ones, even if its ability is removed.
//
// Idol of False Gods: "As long as this artifact has eight or more +1/+1
// counters on it, it's a 0/0 creature in addition to its other types
// and it has annihilator 2." Its 2024-06-07 ruling: "If Idol of False
// Gods loses its abilities before gaining its eighth +1/+1 counter, it
// will have annihilator 2; otherwise, it won't." So the effect is
// ordered by the moment it began to apply, the eighth counter, and not
// by the moment the Idol entered: a removal that arrived in between is
// older than the effect and applies first.
//
// The moment is DATA, because it is history the board cannot recompute:
// Card.CounterStampedAt carries it under a key of its own
// (CounterThreshold.stampKey), written where counters are written
// (applyCounterByLocked) and nowhere else. It rides the snapshot with
// the keyword counters' stamps, and like them it is cleared with the
// counters and on every zone change (CR 400.7). A threshold with no
// stamp — a restore point from before this file, or a card that entered
// with its counters — sorts at its source's own timestamp, the earliest
// moment the effect could have begun.

// CounterThreshold is "N or more counters of Kind" on a static's own
// source. The zero value declares nothing.
type CounterThreshold struct {
	Kind    string
	AtLeast int
}

// IsZero reports whether the threshold declares nothing.
func (t CounterThreshold) IsZero() bool { return t.Kind == "" || t.AtLeast <= 0 }

// stampKey is the threshold's key in Card.CounterStampedAt. A keyword
// counter kind is a bare word ("flying"); this one carries a ">=" and
// a count, so the two can never collide.
func (t CounterThreshold) stampKey() string {
	return t.Kind + ">=" + strconv.Itoa(t.AtLeast)
}

// counterThresholdKinds is the catalog-wide set of counter kinds some
// static declares a threshold on, so a counter placement on any other
// kind pays one map read. Written at boot from effects.Register,
// read-only afterwards — the contract every Catalog* hook carries.
var counterThresholdKinds = struct {
	sync.RWMutex
	kinds map[string]bool
}{kinds: map[string]bool{}}

// IndexCounterThresholds records the counter kinds `statics` declare a
// StartsAtCounters threshold on. Called once per card from
// effects.Register.
func IndexCounterThresholds(statics []StaticAbility) {
	for _, s := range statics {
		if s.StartsAtCounters.IsZero() {
			continue
		}
		counterThresholdKinds.Lock()
		counterThresholdKinds.kinds[s.StartsAtCounters.Kind] = true
		counterThresholdKinds.Unlock()
	}
}

func counterKindHasThreshold(kind string) bool {
	counterThresholdKinds.RLock()
	ok := counterThresholdKinds.kinds[kind]
	counterThresholdKinds.RUnlock()
	return ok
}

// stampCounterThresholdsLocked keeps the threshold stamps of `c` in step
// with a change of its `kind` counters from `before` to `after`: the
// moment the count reaches a declared threshold is recorded, and the
// record goes when the count falls below it again. A count that stays
// above it (a ninth counter) changes nothing, because the effect did
// not begin again.
//
// Caller must hold g.mu.
func stampCounterThresholdsLocked(c *Card, kind string, before, after int, now func() int64) {
	if before == after || !counterKindHasThreshold(kind) {
		return
	}
	for _, s := range staticAbilitiesOf(c) {
		t := s.StartsAtCounters
		if t.IsZero() || t.Kind != kind {
			continue
		}
		key := t.stampKey()
		switch {
		case before < t.AtLeast && after >= t.AtLeast:
			if c.CounterStampedAt == nil {
				c.CounterStampedAt = make(map[string]int64, 1)
			}
			c.CounterStampedAt[key] = now()
		case after < t.AtLeast:
			if _, ok := c.CounterStampedAt[key]; ok {
				delete(c.CounterStampedAt, key)
				if len(c.CounterStampedAt) == 0 {
					c.CounterStampedAt = nil
				}
			}
		}
	}
}

// staticTimestamp is the CR 613.7 timestamp of a static ability of the
// permanent `src`: the permanent's own, or the moment its counters
// reached the static's declared threshold when that is later.
func staticTimestamp(src *Card, ab StaticAbility, own int64) int64 {
	if ab.StartsAtCounters.IsZero() {
		return own
	}
	if at, ok := src.CounterStampedAt[ab.StartsAtCounters.stampKey()]; ok && at > own {
		return at
	}
	return own
}
