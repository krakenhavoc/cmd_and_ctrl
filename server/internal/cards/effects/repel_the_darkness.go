package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Repel the Darkness — Instant {2}{W} (Foundations Jumpstart):
//
//	"Tap up to two target creatures.
//	 Draw a card."
//
// Up to two targets, so it may be cast with none; the draw happens
// either way (CR 608.2c). A target that has gone is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "47f5ea79-2bf2-4e17-9c6d-e1cbcd9d30d1",
		Name:         "Repel the Darkness",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return tapEachLegalTarget(ctx.Game, ctx.Item) }),
			DrawCards{N: 1},
		),
	})
}
