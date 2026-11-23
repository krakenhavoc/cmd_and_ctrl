package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Darksteel Ingot — Artifact {3}:
//
//	"Indestructible (Effects that say "destroy" don't destroy this
//	 artifact.)
//	 {T}: Add one mana of any color."
//
// Thornglint Bridge's indestructibility (PrintedKeywords is what the
// destruction paths read) on a Birds of Paradise mana ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a2529491-7389-4cfa-92d2-145eda779603",
		Name:            "Darksteel Ingot",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		ManaAbilities:   []effects.ManaAbility{effects.Cmdr2B01TapForAnyColor()},
	})
}
