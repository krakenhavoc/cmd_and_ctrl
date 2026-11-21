package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Overgrown Estate — Enchantment {W}{B}{G} (Apocalypse, #5614):
//
//	"Sacrifice a land: You gain 3 life."
//
// Zuran Orb's land sacrifice outlet at 3 life a land. The sacrifice is
// the cost (CR 701.21a), so it can be activated any number of times, at
// instant speed, while you have a land to give up.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4d52c4a5-e5c8-4fb4-be50-78d5482dd1ae",
		Name:         "Overgrown Estate",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a land: You gain 3 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.B08SacrificeALand(),
			Effect:  effects.Do(effects.GainLife{Amount: 3}),
		}},
	})
}
