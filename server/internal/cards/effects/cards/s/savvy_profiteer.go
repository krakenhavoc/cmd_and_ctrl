package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Savvy Profiteer — Creature — Ferengi Rogue {1}{R} (Star Trek, #4973), 2/1:
//
//	"{T}: Add {R}.
//	 {3}, {T}: Create a Treasure token. (It's an artifact with "{T},
//	 Sacrifice this token: Add one mana of any color.")"
//
// Both abilities tap the creature, so summoning sickness bars each the
// turn it arrives (CR 302.6). The Treasure is the printed token row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e96a9f9a-9326-4f0b-98ad-23210c97299f",
		Name:         "Savvy Profiteer",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "{T}: Add {R}.",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Create a Treasure token.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.TreasureToken(), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
