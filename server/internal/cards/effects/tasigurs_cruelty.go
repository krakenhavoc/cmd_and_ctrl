package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tasigur's Cruelty — Sorcery {5}{B}:
//
//	"Delve. Each opponent discards two cards."
//
// Delve (CR 702.66, ADR 0100) is `Delve: true` and nothing else: the
// engine prices it, validates the graveyard cards named on
// `delve_ids` and exiles them at CR 601.2h with the spell on the
// stack.
//
// The APNAP fan-out over every opponent (EachPlayerDiscardsForEffect),
// each discarding player choosing their own two (CR 701.8a). No
// simplification.
func init() {
	Register(Spec{
		OracleID:     "a232daf3-db54-4711-88b0-e5072c05f58e",
		Name:         "Tasigur's Cruelty",
		Completeness: CompletenessFull,
		Delve:        true,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ctx.Game.EachPlayerDiscardsForEffect(ctx.Controller(), game.DiscardPrompt{
				Source: item.SourceCardID,
				N:      2,
			})
			return nil
		},
	})
}
