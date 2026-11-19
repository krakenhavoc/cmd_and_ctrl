package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Star Compass — Artifact {2}:
//
//	"This artifact enters tapped.
//	 {T}: Add one mana of any color that a basic land you control could
//	 produce."
//
// Reflecting Pool's derivation (CR 106.7) over basic lands you control, colours
// only. Like the printed card, it makes nothing without a basic land that
// could produce a colour.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b5e9599f-9561-4e3f-9904-9a4dab98cabb",
		Name:         "Star Compass",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:                    ManaAbilityCost{Tap: true},
			DerivedMatch:            s82Bns04DerivedFromOwnBasicLands(),
			DerivedColorsOnly:       true,
			DerivesFromOtherSources: true,
			Label:                   "Add one mana of any color that a basic land you control could produce",
		}},
	})
}
