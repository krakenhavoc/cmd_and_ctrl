package heuristic

import (
	"context"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
)

// trace.go implements aiseat.Tracer for the heuristic.
//
// The heuristic already has two entry points that answer the same
// question at different resolutions — Decide returns one index, Rank
// returns the prices — and a trace is simply both of them. That is
// worth more than it sounds: on a heuristic seat the decision log
// records the whole ranking, so a reviewer looking at a blunder can
// see the move that was passed over and what the scorer thought it
// was worth, rather than only that it lost.

// Compile-time assertion: the heuristic is a Tracer.
var _ aiseat.Tracer = (*Policy)(nil)

// DecideTraced is Decide with the ranking attached.
//
// It prices the window twice — once in Rank, once inside Decide — and
// that is fine: the model funnel already does exactly this on every
// Layer B window in production (policy.go calls Rank and then
// Fallback.Decide), and both are microseconds against a deadline
// measured in seconds.
//
// Rank and Decide are called SEQUENTIALLY and each takes the policy's
// mutex on its own, exactly as funnel_game_test.go's agreement policy
// does. Holding the lock across both would be the obvious thing and
// would deadlock: both are public entry points that lock for
// themselves.
func (p *Policy) DecideTraced(ctx context.Context, in aiseat.Input) (aiseat.Decision, aiseat.Trace, error) {
	cands := p.Rank(ctx, in)
	d, err := p.Decide(ctx, in)
	tr := aiseat.Trace{Layer: "B", HeuristicIndex: d.Index}
	if err != nil {
		tr.HeuristicIndex = aiseat.Decline
	}
	for _, c := range cands {
		tr.Candidates = append(tr.Candidates, aiseat.Candidate{Index: c.Index, Value: c.Value, Reason: c.Reason})
	}
	return d, tr, err
}
