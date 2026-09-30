package legal

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// either_cost.go — ADR 0100 §6: the bot's announcement policy for an
// either/or additional cost ("sacrifice an artifact or discard a
// card").
//
// One announcement per branch the seat can pay, each priced with its
// own cost_branch through the engine's one pricer, and each paid by the
// existing payment search (discards, sacrifices, blight). A card has at
// most three branches (Dusk Mangler), so this multiplies the expansion
// by a small constant rather than by anything on the board.

// castCostBranches is the branches this seat may announce for `card`:
// [nil] for a card with no either/or cost, which is every card but a
// handful, so the caller's loop runs once and the enumeration is
// unchanged for them. Empty when no branch can be paid — CR 601.2h,
// "Unpayable costs can't be paid" — which offers no cast at all.
//
// Payability is the engine's own predicate
// (game.AdditionalCostBranchPayableLocked), the one CastSpell and the
// view ask, so a bot is never offered a branch the server refuses
// (#544). On top of the rule, the bot policy the alternative cost's
// life already follows: CR 119.4 lets a player pay life down to exactly
// zero, the next state-based check then kills them, and a bot offered
// that line would take it.
func (e *enumerator) castCostBranches(card game.Card) []*int {
	ac := game.AdditionalCostFor(game.CatalogKey(card))
	if !ac.Branched() {
		return []*int{nil}
	}
	var out []*int
	for i := range ac.Either {
		if !e.g.AdditionalCostBranchPayableLocked(e.seat, card, i) {
			continue
		}
		if life := ac.Either[i].PayLife; life > 0 && e.p.Life <= life {
			continue
		}
		b := i
		out = append(out, &b)
	}
	return out
}

// branchLife is the fixed life a chosen branch pays (ADR 0100 §2), 0
// for any other cost. It is part of the move's price the payload names
// only by its branch index, so it rides MoveCost.Life beside the
// alternative cost's life.
func branchLife(c *game.AdditionalCost) int {
	if c == nil {
		return 0
	}
	return c.PayLife
}
