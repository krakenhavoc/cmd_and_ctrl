package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "4d52c4a5-e5c8-4fb4-be50-78d5482dd1ae",
		Name:         "Overgrown Estate",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice a land: You gain 3 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    b08SacrificeALand(),
			Effect:  Do(GainLife{Amount: 3}),
		}},
	})
}
