package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wizards' School — Land (Homelands, #5635):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Add {U}.
//	 {2}, {T}: Add {W} or {B}."
//
// Three mana abilities (CR 605.1a), built by tieredManaLand
// (tiered_mana_lands.go) with the rest of its cycle.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "51635f96-af1d-4b33-9118-432495aaf07f",
		Name:          "Wizards' School",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: effects.TieredManaLand("U", "W", "B"),
	})
}
