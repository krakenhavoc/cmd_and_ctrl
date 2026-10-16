package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Evermind — Instant — Arcane (no mana cost):
//
//	"Draw a card.
//	Splice onto Arcane {1}{U}"
//
// It has no mana cost, so it can't be cast on its own (CR 118.6); the
// engine refuses that cast. Its use is the splice (ADR 0202).
func init() {
	Register(Spec{
		OracleID:     "14e23074-95b8-4477-b981-6231c19bacdc",
		Name:         "Evermind",
		Completeness: CompletenessFull,
		Splice:       SpliceOntoArcane("{1}{U}"),
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
