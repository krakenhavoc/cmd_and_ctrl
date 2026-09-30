package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Treasure Cruise — Sorcery {7}{U}:
//
//	"Delve. Draw three cards."
//
// Delve (CR 702.66, ADR 0100) is `Delve: true` and nothing else: the
// engine prices it, validates the graveyard cards named on
// `delve_ids` and exiles them at CR 601.2h with the spell on the
// stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5b6bdf5a-2742-4851-92cd-a857a3852836",
		Name:         "Treasure Cruise",
		Completeness: CompletenessFull,
		Delve:        true,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DrawCards{Player: ctx.Controller(), N: 3}.Apply(ctx)
		},
	})
}
