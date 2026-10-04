package roadmap

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// declaresExilePermanentCost is the "exiling a permanent you control as
// a cost" probe (#1600): an activated ability or a mana ability whose
// cost carries the exile-a-permanent component
// (game.AbilityCost.ExilePermanents / effects.ManaAbilityCost's). Kept
// out of registry.go so the seam's entry there stays one hunk.
func declaresExilePermanentCost(s effects.Spec) bool {
	for _, a := range s.Activated {
		if !a.Cost.ExilePermanents.Empty() {
			return true
		}
	}
	for _, m := range s.ManaAbilities {
		if !m.Cost.ExilePermanents.Empty() {
			return true
		}
	}
	return false
}
