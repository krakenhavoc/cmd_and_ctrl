package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Infectious Bite — Instant {1}{G}: "Target creature you control deals damage
// equal to its power to target creature you don't control. Each opponent gets
// a poison counter."
//
// Bite Down's one-sided fight (oneSidedBite), then the poison. The poison
// counters are given whether or not the bite still has legal targets, as long
// as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7b740cd6-e07d-4273-8a7f-e084bdaf0f1c",
		Name:         "Infectious Bite",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.OneSidedBite(ctx.Item, ctx) }),
			effects.S82OneB04PoisonStep(),
		),
	})
}
