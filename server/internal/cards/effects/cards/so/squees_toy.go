package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Squee's Toy — Artifact {1} (Tempest, S82 OLD b754, #5628):
//
//	"{T}: Prevent the next 1 damage that would be dealt to target creature
//	 this turn."
//
// Clergy en-Vec's shield, aimed at a creature only.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d0e33a3f-659a-4889-98f6-d0f385e259f1",
		Name:         "Squee's Toy",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Prevent the next 1 damage that would be dealt to target creature this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.PreventNextOneDamageToTheTarget("Squee's Toy"),
		}},
	})
}
