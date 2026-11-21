package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oread of Mountain's Blaze — Enchantment Creature — Nymph {1}{R}, 1/3 (Theros Beyond Death, #5514):
//
//	"{2}{R}, Discard a card: Draw a card."
//
// Famished Foragers' rummage: the discard is part of the cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ebbcbaa4-9dc2-4098-b6a2-393375d86f16",
		Name:         "Oread of Mountain's Blaze",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, Discard a card: Draw a card.",
			Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.DiscardACard()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
