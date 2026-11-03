package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Expansion Algorithm — Sorcery {X}{U}{U} (Secrets of Strixhaven
// Commander, #4834):
//
//	"Proliferate X times. (To proliferate, choose any number of
//	 permanents and/or players, then give each another counter of each
//	 kind already there.)"
//
// Tam, the Possibility's repeated proliferate (proliferateTimesKeyed):
// each proliferate is its own choice (CR 701.34a), asked once the one
// before it is answered, so the player may choose a different set each
// time (the ruling). The repetitions all happen during the resolution,
// with no chance to respond between them. X is the announced value;
// X = 0 proliferates nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "afbc348a-7b27-4577-8c24-cd6d5b894dc1",
		Name:         "Expansion Algorithm",
		XMatters:     true,
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return proliferateTimesKeyed(ctx.Game, ctx.Controller(), ctx.Source(), ctx.X())
		},
	})
}
