package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shivan Gorge — Legendary Land (EDHREC rank 2363):
//
//	"{T}: Add {C}.
//	 {2}{R}, {T}: Shivan Gorge deals 1 damage to each opponent."
//
// A colourless land with a pinger. The mana half is an ordinary tap
// ability; the ping is a CR 602 activated ability sharing the tap,
// so a Gorge tapped for mana cannot ping the same turn. The Gorge
// itself is the damage source, so a damage doubler sees it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e90a1381-c9c1-4f57-928c-5d19dc065274",
		Name:         "Shivan Gorge",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, {T}: Shivan Gorge deals 1 damage to each opponent.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DamageToEachOpponent(g, item, 1)
			},
		}},
	})
}
