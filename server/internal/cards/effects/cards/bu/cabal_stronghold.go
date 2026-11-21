package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cabal Stronghold — Land:
//
//	"{T}: Add {C}.
//	 {3}, {T}: Add {B} for each basic Swamp you control."
//
// Cabal Coffers's scaled mana ability with a {3} cost and a narrower
// count. "Basic Swamp" is the Basic supertype AND the Swamp subtype,
// both read off the EFFECTIVE characteristics, so a nonbasic Swamp
// (Watery Grave) is not counted, and a basic land that Urborg or Blood
// Moon has retyped is counted for what it is now.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "066cd584-773c-4623-be53-8f6feda5a26a",
		Name:         "Cabal Stronghold",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:           effects.ManaAbilityCost{Tap: true, Mana: "{3}"},
				ProducedAmount: effects.ProducedPerPermanent("B", matchBasicSwamp),
				Label:          "{3}, {T}: Add {B} for each basic Swamp you control",
			},
		},
	})
}

func matchBasicSwamp(c game.Card) bool {
	return effects.IsBasicLand(c) && effects.MatchLandSubtype("Swamp")(c)
}
