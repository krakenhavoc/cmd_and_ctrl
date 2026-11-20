package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tower of Eons — Artifact {4} (Mirrodin, #5606):
//
//	"{8}, {T}: You gain 10 life."
//
// One activated row: its controller gains 10 life as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "74a71bbd-c307-481c-b0cf-70e24b0c8ad4",
		Name:         "Tower of Eons",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			mrdB5TowerRow("You gain 10 life.", game.Purpose{Answers: game.AnswerValue}, Do(GainLife{Amount: 10})),
		},
	})
}
