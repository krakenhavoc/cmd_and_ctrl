package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Shire — Legendary Land (EDHREC rank 1165):
//
//	"The Shire enters tapped unless you control a legendary creature.
//	 {T}: Add {G}.
//	 {1}{G}, {T}, Tap an untapped creature you control: Create a Food
//	 token."
//
// Mines of Moria's green sibling: the same "unless you control a
// legendary creature" entry, read post-layer so an animated legendary
// artifact counts, and a plain {G}.
//
// The Food ability is a CR 602 activation with a tap-another cost
// (#1381): Plus(ManaCost("{1}{G}"), TapCost(), TapAnotherUntapped(...)),
// the same TapOthers component the station ability pays one file
// over.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9abf9a0e-8e7d-406b-a01d-d4870b30134e",
		Name:         "The Shire",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTappedUnless(effects.B08ControlsLegendaryCreature)},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}, {T}, Tap an untapped creature you control: Create a Food token.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}{G}"), effects.TapCost(), effects.TapAnotherUntapped("an untapped creature you control", effects.Creature())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{
					Controller: item.Controller,
					Template:   effects.FoodToken(),
					N:          1,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
