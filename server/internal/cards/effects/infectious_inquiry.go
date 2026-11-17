package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Infectious Inquiry — Sorcery {2}{B}: "You draw two cards and you lose 2
// life. Each opponent gets a poison counter." The life loss waits for a prompt
// the draw asks (CR 608.2c). No simplification.
func init() {
	Register(Spec{
		OracleID:     "66051196-53ea-4461-b241-205df9395b3b",
		Name:         "Infectious Inquiry",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx) }),
			Instruction(func(ctx *Context) error { return GainLife{Player: ctx.Controller(), Amount: -2}.Apply(ctx) }),
			s82OneB04PoisonStep(),
		),
	})
}
