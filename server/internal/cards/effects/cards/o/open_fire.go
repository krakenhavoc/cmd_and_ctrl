package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Open Fire — Instant {2}{R}:
//
//	"Open Fire deals 3 damage to any target."
//
// Lightning Strike at three mana: the same any-target clause and three
// damage from the spell as its source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8b218023-50b8-4d9d-b17b-9647479d3675",
		Name:         "Open Fire",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.BnsB47DamageToAnyTarget(3)(ctx.Game, item)
		},
	})
}
