package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hex — Sorcery {4}{B}{B}:
//
//	"Destroy six target creatures."
//
// Exactly six targets, any player's, so it is uncastable with fewer than
// six creatures on the table (CR 601.2c). The six are destroyed
// TOGETHER, as one event (CR 701.8a, 603.10a): a Blood Artist sees all
// six die, and a creature that left in response is skipped rather than
// fizzling the rest (CR 608.2b). Indestructible ones survive, through
// the same guard every destroy takes.
//
// Curtains' Call's loop destroys its targets one after another; this
// card takes Eliminate the Competition's simultaneous path
// (vsDestroyTheTargets), because six leaves-the-battlefield triggers that
// each saw the other creatures already gone (or still there) would be a
// different game.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "89457fa8-b0e2-4277-b6a9-c6601c59b46d",
		Name:         "Hex",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("six target creatures").WithCount(6, 6),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return vsDestroyTheTargets(ctx)
		},
	})
}
