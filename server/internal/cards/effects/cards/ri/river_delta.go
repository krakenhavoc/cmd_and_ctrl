package cardsri

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// River Delta — Land (Ice Age, S82 OLD batch 850, #5636):
//
//	"This land doesn't untap during your untap step if it has a depletion
//	 counter on it.
//	 At the beginning of your upkeep, remove a depletion counter from this land.
//	 {T}: Add {U} or {B}. Put a depletion counter on this land."
//
// Built by S82OldB850DepletionLand (s82_old_b850_helpers.go).
//
// No simplification.
func init() {
	effects.Register(effects.S82OldB850DepletionLand("4c6c064c-da61-4c7a-9607-8fb4490ff9a6", "River Delta", "U", "B"))
}
