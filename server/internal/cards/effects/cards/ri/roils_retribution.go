package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Roil's Retribution — Instant {3}{W}{W} (Battle for Zendikar, #5547):
//
//	"Roil's Retribution deals 5 damage divided as you choose among any number
//	 of target attacking or blocking creatures."
//
// The division is announced with the targets (CR 601.2d) and each target
// must get at least 1; a target that is illegal on resolution is simply
// skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "38e43d06-33f1-4672-b0be-b0ea8bf36291",
		Name:         "Roil's Retribution",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCreature("any number of target attacking or blocking creatures", effects.AttackingOrBlocking()).
			WithCount(0, 0).Dividing(effects.Divide(5)),
		Purpose: game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DealDividedDamage(ctx)
		},
	})
}
