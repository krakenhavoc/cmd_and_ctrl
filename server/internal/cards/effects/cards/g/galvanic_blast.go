package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Galvanic Blast — Instant {R} (Special Guests):
//
//	"Galvanic Blast deals 2 damage to any target.
//	 Metalcraft — Galvanic Blast deals 4 damage instead if you control
//	 three or more artifacts."
//
// Metalcraft is read as the spell resolves (an ability word, no rule of
// its own): the artifacts counted are the ones its controller has then.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d211c03a-03cd-40c5-b7c8-ada352fd01a6",
		Name:         "Galvanic Blast",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 2)), Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := 2
			if effects.ControlsAtLeast(3, effects.MatchArtifact)(ctx.Game, ctx.Controller(), ctx.Source()) {
				n = 4
			}
			return effects.DamageToFirstTargetStep(n)(item, ctx)
		},
	})
}
