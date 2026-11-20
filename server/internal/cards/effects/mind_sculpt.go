package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mind Sculpt — Sorcery {1}{U} (Magic 2013, #5568):
//
//	"Target opponent mills seven cards."
//
// Dampen Thought's mill at seven, aimed at an opponent only.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4d8592d7-2e36-4a7a-be9e-7c8f512d5f62",
		Name:         "Mind Sculpt",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return TargetPlayerMills(7)(ctx.Game, item)
		},
	})
}
