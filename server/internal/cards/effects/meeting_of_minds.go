package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Meeting of Minds — Instant {3}{U} (March of the Machine): "Convoke. Draw
// two cards."
func init() {
	Register(Spec{
		OracleID:     "3b91dc3e-46ef-4a68-829f-ef7d79598cf7",
		Name:         "Meeting of Minds",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Purpose:      game.Purpose{Draws: 2},
		OnResolve:    ResolveSteps(Instruction(func(ctx *Context) error { return DrawCards{N: 2}.Apply(ctx) })),
	})
}
