package rules

import (
	"context"
	"sort"
	"sync"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
)

// meter.go is the instrumentation half of Layer A, and it exists
// because ADR 0033 §5 states an acceptance number rather than a
// preference: "if Layer A is not absorbing >80% of windows, the
// funnel is broken and should be fixed before reaching for a cheaper
// model." A number that is never measured is a wish, so the filter
// counts itself and the count is what the sprint reports.

// Meter counts windows and how many of them Layer A absorbed. The
// zero value is ready to use and safe from several goroutines, which
// it has to be: one Meter is normally shared by the four runners at a
// table so the rate is a property of the game rather than of a seat.
type Meter struct {
	mu       sync.Mutex
	windows  int64
	absorbed int64
	byRule   map[string]int64
}

// Observe records one verdict.
func (m *Meter) Observe(v Verdict) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.windows++
	if v.Absorbed() {
		m.absorbed++
	}
	if m.byRule == nil {
		m.byRule = map[string]int64{}
	}
	rule := v.Rule
	if rule == "" {
		rule = RuleNone
	}
	m.byRule[rule]++
}

// MeterStats is a snapshot of a Meter.
type MeterStats struct {
	// Windows is every decision window the filter saw.
	Windows int64
	// Absorbed is how many of them it resolved with no further layer.
	Absorbed int64
	// ByRule counts every window by the rule that fired, escalations
	// included (under RuleNone).
	ByRule map[string]int64
}

// Rate is the absorption rate in [0,1]. Zero windows reads as zero
// rather than as one: a filter that has seen nothing has proved
// nothing.
func (s MeterStats) Rate() float64 {
	if s.Windows == 0 {
		return 0
	}
	return float64(s.Absorbed) / float64(s.Windows)
}

// Rules returns the rule names seen, most frequent first, so a log
// line can say WHICH rule is carrying the rate.
func (s MeterStats) Rules() []string {
	out := make([]string, 0, len(s.ByRule))
	for k := range s.ByRule {
		out = append(out, k)
	}
	sort.Slice(out, func(i, j int) bool {
		if s.ByRule[out[i]] != s.ByRule[out[j]] {
			return s.ByRule[out[i]] > s.ByRule[out[j]]
		}
		return out[i] < out[j]
	})
	return out
}

// Stats snapshots the meter.
func (m *Meter) Stats() MeterStats {
	if m == nil {
		return MeterStats{}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	by := make(map[string]int64, len(m.byRule))
	for k, v := range m.byRule {
		by[k] = v
	}
	return MeterStats{Windows: m.windows, Absorbed: m.absorbed, ByRule: by}
}

// Reset zeroes the meter.
func (m *Meter) Reset() {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.windows, m.absorbed, m.byRule = 0, 0, nil
}

// Filter is Layer A in front of any policy: it answers the windows
// the rules settle and delegates the rest. Wrapping the heuristic
// with it IS ADR 0033 §6's `heuristic` tier, which the tier table
// defines as "A + B".
//
// Filter forwards the aiseat.Conceder question untouched. Conceding
// is a judgement about the position and Layer A holds no opinions
// about positions.
type Filter struct {
	// Inner is the policy that gets the windows Layer A will not
	// answer. Required.
	Inner aiseat.Policy
	// Meter is optional; nil disables counting.
	Meter *Meter
	// Tier overrides the reported policy name. Empty reports
	// "rules+<inner>".
	Tier string
}

// NewFilter wraps inner with the Layer A filter and the given meter.
func NewFilter(inner aiseat.Policy, m *Meter) *Filter {
	return &Filter{Inner: inner, Meter: m}
}

// Name is the tier name.
func (f *Filter) Name() string {
	if f.Tier != "" {
		return f.Tier
	}
	return "rules+" + f.Inner.Name()
}

// Decide runs Layer A and delegates on escalation.
func (f *Filter) Decide(ctx context.Context, in aiseat.Input) (aiseat.Decision, error) {
	v := Resolve(in)
	f.Meter.Observe(v)
	if v.Absorbed() {
		return aiseat.Decision{Index: v.Index, Reason: v.Reason}, nil
	}
	return f.Inner.Decide(ctx, in)
}

// ShouldConcede forwards to the inner policy when it is a Conceder.
func (f *Filter) ShouldConcede(in aiseat.Input) bool {
	c, ok := f.Inner.(aiseat.Conceder)
	return ok && c.ShouldConcede(in)
}

// Compile-time assertion: the filter is a Tracer.
var _ aiseat.Tracer = (*Filter)(nil)

// DecideTraced is Decide with the verdict shown: Layer A's own answer
// when it absorbed the window, and otherwise whatever the inner
// policy has to say about it.
//
// It is a SEPARATE path from Decide rather than a wrapper around it,
// because both run Resolve and both tick the meter: calling one from
// the other would double-count every window the absorption rate is
// computed from, and that rate is ADR 0033 §5's acceptance number.
// The runner calls one or the other, never both.
func (f *Filter) DecideTraced(ctx context.Context, in aiseat.Input) (aiseat.Decision, aiseat.Trace, error) {
	v := Resolve(in)
	f.Meter.Observe(v)
	if v.Absorbed() {
		// HeuristicIndex stays Decline: nobody asked Layer B, and
		// recording 0 here would read as "the heuristic wanted the
		// first move", which is a different and false claim.
		return aiseat.Decision{Index: v.Index, Reason: v.Reason},
			aiseat.Trace{Layer: "A", Rule: v.Rule, HeuristicIndex: aiseat.Decline}, nil
	}
	if t, ok := f.Inner.(aiseat.Tracer); ok {
		return t.DecideTraced(ctx, in)
	}
	d, err := f.Inner.Decide(ctx, in)
	tr := aiseat.Trace{Layer: "B", HeuristicIndex: d.Index}
	if err != nil {
		tr.HeuristicIndex = aiseat.Decline
	}
	return d, tr, err
}
