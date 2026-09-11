package heuristic

import (
	"context"
	"sort"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// rank.go exposes the scorer's working, which Layer C needs and
// Decide throws away.
//
// ADR 0033 §5 lists "top-two heuristic candidates within ε" as an
// escalation trigger, glossed as "the heuristic is admitting it does
// not know". That is a genuinely good signal and it is the only
// trigger that cannot be read off the board: it is a property of the
// SCORES, and Decide returns one index. Rank returns the prices so
// the layer above can see the margin.
//
// It is a new entry point rather than a change to Decide because the
// heuristic must keep standing alone (ADR 0033 §6): nothing here
// alters what Decide answers, and a policy that never calls Rank is
// the `heuristic` tier exactly as it shipped.

// Candidate is one move with the price the heuristic put on it.
// Value is in the same units score.go uses; passing is the zero.
type Candidate struct {
	Index  int
	Value  float64
	Reason string
}

// Rank prices every move in the window, best first.
//
// It returns nil for the windows the scalar scorer does not answer —
// mulligans, and any window containing a combat declaration. Those
// are planned in decideMulligan and combat.go rather than priced, so
// a "ranking" of them would be a fiction, and a caller comparing the
// top two of a fiction would escalate on noise. Combat has its own
// escalation trigger and does not need this one.
//
// Rank respects ctx the same way decideGeneral does: it returns what
// it has priced so far rather than blowing a deadline.
func (p *Policy) Rank(ctx context.Context, in aiseat.Input) []Candidate {
	if len(in.Moves) == 0 {
		return nil
	}
	if allKind(in.Moves, legal.KindMulligan) ||
		anyKind(in.Moves, legal.KindBlock) || anyKind(in.Moves, legal.KindAttack) {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	st := p.newState(in)
	choices := allKind(in.Moves, legal.KindChoice)

	out := make([]Candidate, 0, len(in.Moves))
	for i := range in.Moves {
		if i%16 == 0 && ctx.Err() != nil {
			break
		}
		var v float64
		var reason string
		if choices {
			v, reason = p.valueOfChoice(st, in.Moves[i])
		} else {
			v, reason = p.valueOf(st, in.Moves[i])
		}
		out = append(out, Candidate{Index: i, Value: v, Reason: reason})
	}
	// Descending by value; ties break to the lower index, which is
	// the enumerator's own preference order and what decideChoice
	// already does.
	sort.SliceStable(out, func(i, j int) bool { return out[i].Value > out[j].Value })
	return out
}
