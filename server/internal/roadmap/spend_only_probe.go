package roadmap

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// declaresSpendOnly is the "costs that only some mana can pay" probe
// (#1600): an activated ability whose cost carries a "spend only
// <colour> mana" clause (game.AbilityCost.SpendOnly), or a mana ability
// whose mana is restricted to monocolored or multicolored spells. Kept
// out of registry.go so the seam's entry there stays one hunk.
func declaresSpendOnly(s effects.Spec) bool {
	for _, a := range s.Activated {
		if a.Cost.SpendOnly != nil {
			return true
		}
	}
	for _, m := range s.ManaAbilities {
		if slices.Contains(m.Restrictions, game.ManaRestrictMulticolored) || slices.Contains(m.Restrictions, game.ManaRestrictMonocolored) {
			return true
		}
	}
	return false
}
