package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Downpour — Instant {1}{U} (Magic 2013, #5568):
//
//	"Tap up to three target creatures."
//
// Zero to three creature targets; each one still legal as it resolves
// is tapped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "78eb5fb1-cb91-4300-9056-4225a416aec0",
		Name:         "Downpour",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("up to three target creatures").WithCount(0, 3),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return tapEachLegalTarget(ctx.Game, item)
		},
	})
}
