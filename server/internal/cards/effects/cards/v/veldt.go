package cardsv

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Veldt — Land (Ice Age, S82 OLD batch 850, #5636):
//
//	"This land doesn't untap during your untap step if it has a depletion
//	 counter on it.
//	 At the beginning of your upkeep, remove a depletion counter from this land.
//	 {T}: Add {G} or {W}. Put a depletion counter on this land."
//
// Built by S82OldB850DepletionLand (s82_old_b850_helpers.go).
//
// No simplification.
func init() {
	effects.Register(effects.S82OldB850DepletionLand("10b9bfb3-c478-45c6-b227-9c66b63bc79b", "Veldt", "G", "W"))
}
