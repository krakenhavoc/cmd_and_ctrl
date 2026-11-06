package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mind Spring — Sorcery {X}{U}{U} (Bloomburrow, #4467; first printed in
// Mirage):
//
//	"Draw X cards."
//
// X is announced with the cast (CR 107.3a, CR 601.2b) and read back as
// the spell resolves; the X cards are drawn one at a time (CR 121.2).
// XMatters keeps the bot from offering the X = 0 cast that does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3428287b-cfd5-45bc-b5f9-1e3f5a425a68",
		Name:         "Mind Spring",
		Completeness: CompletenessFull,
		XMatters:     true,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DrawCards{N: ctx.X()}.Apply(ctx)
		},
	})
}
