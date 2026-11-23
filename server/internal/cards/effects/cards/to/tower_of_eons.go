package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tower of Eons — Artifact {4} (Mirrodin, #5606):
//
//	"{8}, {T}: You gain 10 life."
//
// One activated row: its controller gains 10 life as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "74a71bbd-c307-481c-b0cf-70e24b0c8ad4",
		Name:         "Tower of Eons",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.MrdB5TowerRow("You gain 10 life.", game.Purpose{Answers: game.AnswerValue}, effects.Do(effects.GainLife{Amount: 10})),
		},
	})
}
