package legal

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// variable_sacrifice_cost.go — ADR 0100 §6: the bot's payments for a
// VARIABLE sacrifice as a cast's additional cost. "Sacrifice X lands"
// (Devastating Summons), "sacrifice any number of creatures" (Vicious
// Betrayal), "you may sacrifice any number of creatures" (Torgaar).

// castVariableSacrificePayments turns the clause's candidates, already
// in payment order, into the payments a cast is offered with:
//
//   - ZERO, when the clause's floor allows it — "any number" does, and
//     so does "sacrifice X" when the card's X is not the whole of its
//     effect (enumeratedXFloor, #810: a Devastating Summons at X = 0
//     makes two 0/0s, which is no move worth offering);
//   - then up to maxEnumeratedVariableCounts positive counts, smallest
//     first, each a prefix of `ordered`, so the counts nest and a bot
//     asked to sacrifice three eats the two it would have eaten for two.
//
// Small first for the reason variableSacrificePayments gives for an
// ability: a count in a COST must not become an arity of the target
// cross product (ADR 0033 §1), and the smallest counts spend the least
// board. Each payment is priced on its own by the caller, so a
// per-sacrifice discount is never advertised at a price the engine
// will not charge.
//
// Every count is checked with game.SacrificeCountLegal, the predicate
// the validator enforces, at the X the count announces for the X form
// (CR 107.3i). Nil when nothing can be offered — a sacrifice-X card
// whose floor the board cannot reach.
func castVariableSacrificePayments(ordered []uuid.UUID, spec *game.TargetSpec, xFloor int) [][]uuid.UUID {
	fromX := game.SacrificeCountFromX(spec)
	lo := xFloor
	if !fromX {
		lo, _ = game.SacrificeCostBounds(spec, 0)
	}
	if lo < 0 {
		lo = 0
	}
	var out [][]uuid.UUID
	if lo == 0 {
		if game.SacrificeCountLegal(spec, 0, 0) {
			out = append(out, nil)
		}
		lo = 1
	}
	positive := 0
	for n := lo; n <= len(ordered) && positive < maxEnumeratedVariableCounts; n++ {
		x := 0
		if fromX {
			x = n
		}
		if !game.SacrificeCountLegal(spec, x, n) {
			break
		}
		out = append(out, ordered[:n])
		positive++
	}
	return out
}
