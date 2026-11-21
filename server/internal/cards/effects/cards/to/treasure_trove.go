package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Treasure Trove — Enchantment {2}{U}{U} (Exodus, #5626):
//
//	"{2}{U}{U}: Draw a card."
//
// One activated ability with a mana cost and the shared draw step. No
// tap and no timing restriction, so it can be activated any time its
// controller has priority, as often as they have the mana.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3c4a6b34-a9bb-426d-908c-7ae0011591df",
		Name:         "Treasure Trove",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}{U}: Draw a card.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}{U}{U}"),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
