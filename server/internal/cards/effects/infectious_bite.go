package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "7b740cd6-e07d-4273-8a7f-e084bdaf0f1c",
		Name:         "Infectious Bite",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature you don't control", OpponentControls()),
		),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return oneSidedBite(ctx.Item, ctx) }),
			s82OneB04PoisonStep(),
		),
	})
}
