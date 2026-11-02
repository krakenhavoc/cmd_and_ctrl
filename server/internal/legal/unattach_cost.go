package legal

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// unattach_cost.go — #4799: the enumerator's half of "Unattach an
// Equipment from <this>" (game.AbilityCost.UnattachEquipment, Captain
// America, First Avenger).
//
// The pick is a dimension of the ROW rather than of the payment loops,
// because a divided amount may read it (game.DivideAmountArgs.Unattached):
// the announcement's steps are bound before the payments are solved, so
// the Equipment has to be known by then. Each candidate is therefore its
// own row, exactly as an either/or cost's branches are (expandCostBranches),
// and a row whose cost has the component and no Equipment attached offers
// no move (CR 118.3, #544).

// expandUnattachPicks returns `rows` with every row whose cost unattaches
// an Equipment replaced by one row per Equipment attached to `sourceID`
// (game.UnattachCostOptionsForEffect, the walk the engine validates
// against). effects.Register admits only a count of 1, so one Equipment
// is one payment.
func (e *enumerator) expandUnattachPicks(sourceID uuid.UUID, rows []costBranchRow) []costBranchRow {
	out := rows[:0:0]
	var options []uuid.UUID
	looked := false
	for _, r := range rows {
		if r.ab.Cost.UnattachEquipment <= 0 {
			out = append(out, r)
			continue
		}
		if !looked {
			options, looked = e.g.UnattachCostOptionsForEffect(sourceID), true
		}
		for _, id := range options {
			pick := r
			pick.unattach = []uuid.UUID{id}
			out = append(out, pick)
		}
	}
	return out
}

// unattachLabel names the Equipment a move unattaches, for the move's
// label: " (unattaching Shield)".
func unattachLabel(g *game.Game, ids []uuid.UUID) string {
	if len(ids) == 0 {
		return ""
	}
	names := make([]string, len(ids))
	for i, id := range ids {
		names[i] = cardName(g, id)
	}
	return " (unattaching " + strings.Join(names, ", ") + ")"
}
