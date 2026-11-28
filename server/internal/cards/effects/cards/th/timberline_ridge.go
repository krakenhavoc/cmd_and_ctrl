package cardsth

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Timberline Ridge — Land (Ice Age, S82 OLD batch 850, #5636):
//
//	"This land doesn't untap during your untap step if it has a depletion
//	 counter on it.
//	 At the beginning of your upkeep, remove a depletion counter from this land.
//	 {T}: Add {R} or {G}. Put a depletion counter on this land."
//
// Built by S82OldB850DepletionLand (s82_old_b850_helpers.go).
//
// No simplification.
func init() {
	effects.Register(effects.S82OldB850DepletionLand("a597d2b7-484b-4cfd-89a6-d166cb1a3420", "Timberline Ridge", "R", "G"))
}
