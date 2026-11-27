package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vanquish the Foul — Sorcery {5}{W} (Theros, S82 OLD b392, #5561):
//
//	"Destroy target creature with power 4 or greater. Scry 1."
//
// The power bound is judged at announce and again as the spell resolves
// (CR 608.2b), on the creature's current power (PowerGE); the scry follows
// the destruction.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6990ac82-49d4-412d-9c3d-1e7fb5479313",
		Name:         "Vanquish the Foul",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Scry: 1},
		Targets:      effects.TargetCreature("target creature with power 4 or greater", effects.PowerGE(4)),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.DestroyTheTargetPermanent(ctx.Item, ctx) }),
			effects.Instruction(func(ctx *effects.Context) error { return effects.Scry{Player: ctx.Controller(), N: 1}.Apply(ctx) }),
		),
	})
}
