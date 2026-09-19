package aiseat

import "time"

// stats_accessor.go is #505 part 2: a path from a Runner to what the
// MODEL FUNNEL underneath it reports — the layer/escalation/token
// counters model.Policy has counted since sub-PR 7
// (aiseat/model/metrics.go) — without this package importing
// aiseat/model.
//
// It cannot import it: aiseat/model already imports aiseat, for
// aiseat.Policy, aiseat.TokenUsage and aiseat.Spend (see spend.go and
// model.Policy.Spend), so the reverse import would be a cycle. The
// fix is the one this package already uses for Spend — a
// package-neutral projection type declared HERE, and a method on
// model.Policy (in the model package, which may reference this one)
// that fills it in.
//
// A `random` or `heuristic` seat has no model.Policy in its wrapper
// chain at all, so reporting this is optional, exactly like Spender,
// and found the same way: aiseat.Capability walks the wrapper chain
// and returns the outermost implementer.

// PolicyStats is the model funnel's per-decision instrumentation,
// projected into the shape a Runner can carry without this package
// importing aiseat/model. It mirrors model.Stats field for field —
// see aiseat/model/metrics.go for what each counter means, and
// model.Policy.PolicyStats for the projection that builds one.
//
// JSON-tagged because this is #505's whole point: the numbers leave
// the process, through GET /games/{id}/bot/stats (server/internal/
// lobby/botstats.go).
type PolicyStats struct {
	// Windows is every Decide call the funnel answered.
	Windows int64 `json:"windows"`
	// ByLayer counts windows by the layer that answered them: "A",
	// "B" or "C" — see aiseat/model's Layer* constants.
	ByLayer map[string]int64 `json:"by_layer,omitempty"`
	// ByEscalation counts how many windows each trigger fired on. A
	// window with two triggers counts in both.
	ByEscalation map[string]int64 `json:"by_escalation,omitempty"`
	// ByFallback counts the reasons Layer C's answer was discarded.
	ByFallback map[string]int64 `json:"by_fallback,omitempty"`
	// Escalated counts windows where at least one trigger fired,
	// once per window — unlike ByEscalation, which counts every
	// trigger a window fired.
	Escalated int64 `json:"escalated"`
	// ModelCalls is how many decision calls were actually attempted.
	ModelCalls int64 `json:"model_calls"`
	// ModelTimeouts is how many of them ran out of budget rather
	// than failing outright — watch this next to MaxModelLatency.
	ModelTimeouts int64 `json:"model_timeouts"`
	// Usage totals every decision call's reported tokens.
	Usage TokenUsage `json:"usage"`
	// ModelLatency totals the time spent inside decision calls;
	// MaxModelLatency is the worst one.
	ModelLatency    time.Duration `json:"model_latency_ns,omitempty"`
	MaxModelLatency time.Duration `json:"max_model_latency_ns,omitempty"`
	// ByImprov counts ADR 0033 §8 improvisation windows by outcome.
	ByImprov map[string]int64 `json:"by_improv,omitempty"`
	// ImprovCalls and ImprovUsage are the improvisation half of the
	// spend, kept apart from ModelCalls/Usage for the same reason
	// aiseat.Spend keeps Decision and Improvisation apart — see
	// spend.go.
	ImprovCalls   int64         `json:"improv_calls"`
	ImprovUsage   TokenUsage    `json:"improv_usage"`
	ImprovLatency time.Duration `json:"improv_latency_ns,omitempty"`
}

// PolicyStatser is an optional Policy extension: a policy that
// funnels decisions through a model reports the funnel's own
// instrumentation. A `random` or `heuristic` seat is not one and
// contributes nothing, which is the true answer for a seat with no
// model.Policy underneath it.
//
// Found through aiseat.Capability, the same mechanism every other
// optional Policy extension uses (see capability.go), so a policy
// wrapped by a test harness or a future tier still reports it as
// long as the wrapper declares Unwrap.
type PolicyStatser interface {
	PolicyStats() PolicyStats
}

// PolicyStats returns the outermost model funnel's instrumentation
// in r's policy wrapper chain, and false when there is none — every
// `random` and `heuristic` seat, and an `assisted` or `strong` seat
// wrapped by something that does not forward it.
//
// It is the accessor #505 part 2 asks for. Runner.Stats() already
// carries this seat's own counters (decisions, applied, latency
// percentiles, spend); PolicyStats is the funnel's separate
// layer/escalation counters, kept apart rather than folded into
// Stats because they only exist for a seat with a model underneath
// it, and a caller asking for them already gets the right zero value
// for free from the bool this returns.
func (r *Runner) PolicyStats() (PolicyStats, bool) {
	return Capability[PolicyStatser](r.policy)
}
