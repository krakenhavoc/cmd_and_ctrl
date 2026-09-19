package heuristic

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// target_order.go implements aiseat.TargetOrderer: the heuristic
// decides which of a spell's candidate targets survive the
// enumerator's expansion cap (#687, ADR 0033 §1).
//
// The cap (legal.Options.MaxExpansionPerSource, 12) is spent in
// candidate order, so before this a removal spell pointed at a board
// of twenty permanents was offered the first twelve the engine
// happened to walk — and the table leader's best creature could
// simply be absent from the move list. A policy cannot pick a move it
// was never offered, so no amount of scoring one layer up could fix
// it.
//
// It reuses THIS package's scoring rather than adding a second one,
// which is the whole reason the ordering is injected from up here
// instead of living in `legal`:
//
//   - a seat is priced by Weights.Threat, the same function the
//     attack rotation ranks opponents with (threat.go);
//   - a permanent is priced by Weights.boardValue, the same function
//     Evaluate adds up, so #727's attachment roles and every
//     restriction discount apply here too.
//
// Two things it deliberately does NOT do.
//
// It does not know what the SPELL is, so it cannot prefer an
// opponent's creature for a Murder and its own for a Giant Growth.
// The ordering is by IMPORTANCE — the biggest objects on the table
// survive the cap, whoever controls them — and choosing among the
// survivors stays Decide's job, where the spell is known. That is the
// right split: this function's only power is to stop a target being
// dropped, and dropping the board's biggest permanent is wrong for
// every spell.
//
// It does not price anything off the battlefield. A spell on the
// stack, a card in a graveyard and a planeswalker's controller all
// score zero, which keeps them in the engine's own order. Those
// clauses have a handful of candidates and the cap never bites on
// them; the day one does, permanentValue is not the function to reach
// for.

// TargetOrder implements aiseat.TargetOrderer. `in` carries the View
// and the Seat; Moves is empty, because this is called in order to
// build them.
//
// One evaluation of the board per decision, captured by the closure,
// so the enumerator's per-candidate call is a map lookup and some
// arithmetic rather than a re-scan of the battlefield.
func (p *Policy) TargetOrder(in aiseat.Input) legal.TargetOrder {
	p.mu.Lock()
	st := p.newState(in)
	p.mu.Unlock()
	return func(c legal.TargetCandidate) float64 {
		id := c.ID.String()
		if c.Player {
			return st.w.Threat(st.evals[id])
		}
		if card := st.bf[id]; card != nil {
			return st.w.boardValue(card, st.attach)
		}
		return 0
	}
}

// compile-time proof that the heuristic really is a TargetOrderer.
// Without it the runner's type assertion would silently answer false
// after a signature change and the ordering would quietly stop
// happening — which is exactly the failure #687 describes, restored.
var _ aiseat.TargetOrderer = (*Policy)(nil)
