package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Duskmantle, House of Shadow — Land (The List):
//
//	"{T}: Add {C}.
//	 {U}{B}, {T}: Target player mills a card."
//
// The colourless ability is a mana ability; the mill targets (any player,
// the controller included) and is an ordinary activated ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "67b2cd0c-ecc8-4129-b1ac-820c9924190c",
		Name:         "Duskmantle, House of Shadow",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}{B}, {T}: Target player mills a card.",
			Cost:    effects.Plus(effects.ManaCost("{U}{B}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.S82BNSB08MillTargetPlayer,
		}},
	})
}
