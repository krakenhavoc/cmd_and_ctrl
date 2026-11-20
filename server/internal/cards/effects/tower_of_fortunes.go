package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tower of Fortunes — Artifact {4} (Mirrodin, #5606):
//
//	"{8}, {T}: Draw four cards."
//
// One activated row: its controller draws four cards as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dfb23df9-f912-4daa-9842-01bd12b4a72a",
		Name:         "Tower of Fortunes",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			mrdB5TowerRow("Draw four cards.", game.Purpose{Answers: game.AnswerValue, Draws: 4}, Do(DrawCards{N: 4})),
		},
	})
}
