package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Open Fire — Instant {2}{R}:
//
//	"Open Fire deals 3 damage to any target."
//
// Lightning Strike at three mana: the same any-target clause and three
// damage from the spell as its source.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8b218023-50b8-4d9d-b17b-9647479d3675",
		Name:         "Open Fire",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return bnsB47DamageToAnyTarget(3)(ctx.Game, item)
		},
	})
}
