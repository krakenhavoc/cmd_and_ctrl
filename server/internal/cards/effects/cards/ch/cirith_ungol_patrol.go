package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cirith Ungol Patrol — Creature — Orc Soldier {4}{B}, 4/5 (The Lord of the
// Rings: Tales of Middle-earth):
//
//	"{1}, {T}, Sacrifice another creature: Draw a card, then create a Food
//	 token."
//
// Agency Coroner's sacrifice outlet (SacrificeAnotherN, "another" as an
// object) with a draw and then a Food, in the printed order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c9610d33-4cef-4171-bbf4-48d720492d55",
		Name:         "Cirith Ungol Patrol",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice another creature: Draw a card, then create a Food token.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, Draws: 1, Tokens: 1},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Effect:  effects.Do(effects.DrawCards{N: 1}, effects.CreateToken{Template: effects.FoodToken(), N: 1}),
		}},
	})
}
