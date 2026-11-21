package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tower of Fortunes — Artifact {4} (Mirrodin, #5606):
//
//	"{8}, {T}: Draw four cards."
//
// One activated row: its controller draws four cards as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dfb23df9-f912-4daa-9842-01bd12b4a72a",
		Name:         "Tower of Fortunes",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.MrdB5TowerRow("Draw four cards.", game.Purpose{Answers: game.AnswerValue, Draws: 4}, effects.Do(effects.DrawCards{N: 4})),
		},
	})
}
