package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Harvester Druid — Creature — Human Druid {1}{G}, 1/1 (Judgment, #5611):
//
//	"{T}: Add one mana of any color that a land you control could produce."
//
// Naga Vitalist's derived mana ability, colours only (colourless is a type, not
// a colour): what the lands you control could produce is read live (CR 106.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2b4f43a-6609-43ca-ad0d-897b6f8cb22f",
		Name:         "Harvester Druid",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:                    effects.ManaAbilityCost{Tap: true},
			DerivedMatch:            effects.DerivedFromOwnLands(),
			DerivedColorsOnly:       true,
			DerivesFromOtherSources: true,
			Label:                   "Add one mana of any color that a land you control could produce",
		}},
	})
}
