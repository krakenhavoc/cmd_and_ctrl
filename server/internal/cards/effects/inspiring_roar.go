package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inspiring Roar — Sorcery {3}{W} (Aether Revolt,
// #5536):
//
//	"Put a +1/+1 counter on each creature you control."
//
// Basri's Solidarity.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "96cf2810-35b8-437e-b609-f9765604b8cf",
		Name:         "Inspiring Roar",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return b11PutCounterOnEachCreatureYouControl(ctx.Game, item)
		},
	})
}
