package legal

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blight_ability_cost.go — #4684: the enumerator's half of "Blight N" as
// part of an activated ability's cost (game.AbilityCost.Blight, Gristle
// Glutton).
//
// One row per creature the seat could blight, rather than the cast
// path's one cheapest payment (blightPayment): which creature takes the
// counters is the whole price of the activation, and the policy can
// read it — a 1/1 token that dies of them, the source itself, or a big
// body that shrugs off a point. The rows come in blightPayment's order,
// a creature that survives the counters first and among those the least
// powerful, so a policy that takes the first row pays the cheapest way.
// The candidates are the engine's BlightOptionsForEffect, so a row whose
// cost blights and has nothing to blight offers no move (CR 701.68b, CR
// 118.3, #544).

// expandBlightPicks returns `rows` with every row whose cost blights
// replaced by one row per creature the seat could put the counters on.
func (e *enumerator) expandBlightPicks(rows []costBranchRow) []costBranchRow {
	out := rows[:0:0]
	for _, r := range rows {
		n := r.ab.Cost.Blight
		if n <= 0 {
			out = append(out, r)
			continue
		}
		for _, id := range e.blightCandidates(n, nil) {
			pick := r
			pick.blight = []uuid.UUID{id}
			out = append(out, pick)
		}
	}
	return out
}

// blightLabel names the creature a move blights, for the move's label:
// " (blighting Gristle Glutton)".
func blightLabel(g *game.Game, ids []uuid.UUID) string {
	if len(ids) == 0 {
		return ""
	}
	names := make([]string, len(ids))
	for i, id := range ids {
		names[i] = cardName(g, id)
	}
	return " (blighting " + strings.Join(names, ", ") + ")"
}
