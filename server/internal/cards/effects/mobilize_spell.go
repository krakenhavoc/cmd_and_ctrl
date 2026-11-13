package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mobilize — Sorcery {G} (Portal, #5630):
//
//	"Untap all creatures you control."
//
// Vitalize's body as a sorcery (b34UntapAllCreaturesYouControl).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7416c8db-ec33-4d92-866c-b7fdc69c8b41",
		Name:         "Mobilize",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return b34UntapAllCreaturesYouControl(ctx.Game, item)
		},
	})
}
