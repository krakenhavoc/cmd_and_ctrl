package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lock and Load — Sorcery {2}{U} (Outlaws of Thunder Junction Commander):
//
//	"Draw a card, then draw a card for each other instant and sorcery spell
//	 you've cast this turn.
//	 Plot {3}{U} (...)"
//
// The turn's cast tally counts Lock and Load itself (it is a sorcery that
// was cast, from the hand or from exile, before it resolves), so "each
// other" is the tally less one, and the draw is that many plus the first.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "9355db6e-d8eb-43be-9c1d-04f1806dc366",
		Name:           "Lock and Load",
		Completeness:   CompletenessFull,
		Purpose:        game.Purpose{Draws: 1},
		SpecialActions: []game.SpecialAction{Plot("{3}{U}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			others := max(0, ctx.Game.CastTallyFor(item.Controller).InstantSorcery-1)
			return DrawCards{Player: item.Controller, N: 1 + others}.Apply(ctx)
		},
	})
}
