package cardskr

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Lava Tubes — Land (Ice Age, S82 OLD batch 850, #5636):
//
//	"This land doesn't untap during your untap step if it has a depletion
//	 counter on it.
//	 At the beginning of your upkeep, remove a depletion counter from this land.
//	 {T}: Add {B} or {R}. Put a depletion counter on this land."
//
// Built by S82OldB850DepletionLand (s82_old_b850_helpers.go).
//
// No simplification.
func init() {
	effects.Register(effects.S82OldB850DepletionLand("2ffb2647-60bb-4916-b919-cbcf18e5e424", "Lava Tubes", "B", "R"))
}
