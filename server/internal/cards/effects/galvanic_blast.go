package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Galvanic Blast — Instant {R} (Special Guests):
//
//	"Galvanic Blast deals 2 damage to any target.
//	 Metalcraft — Galvanic Blast deals 4 damage instead if you control
//	 three or more artifacts."
//
// Metalcraft is read as the spell resolves (an ability word, no rule of
// its own): the artifacts counted are the ones its controller has then.
func init() {
	Register(Spec{
		OracleID:     "d211c03a-03cd-40c5-b7c8-ada352fd01a6",
		Name:         "Galvanic Blast",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 2)), Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			n := 2
			if ControlsAtLeast(3, MatchArtifact)(ctx.Game, ctx.Controller(), ctx.Source()) {
				n = 4
			}
			return damageToFirstTarget(n)(item, ctx)
		},
	})
}
