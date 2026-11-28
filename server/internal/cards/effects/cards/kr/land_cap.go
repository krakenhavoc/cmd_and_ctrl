package cardskr

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Land Cap — Land (Ice Age, S82 OLD batch 850, #5636):
//
//	"This land doesn't untap during your untap step if it has a depletion
//	 counter on it.
//	 At the beginning of your upkeep, remove a depletion counter from this land.
//	 {T}: Add {W} or {U}. Put a depletion counter on this land."
//
// Built by S82OldB850DepletionLand (s82_old_b850_helpers.go).
//
// No simplification.
func init() {
	effects.Register(effects.S82OldB850DepletionLand("bfec4d0a-3792-4bc3-bae1-e639da5bb9a6", "Land Cap", "W", "U"))
}
