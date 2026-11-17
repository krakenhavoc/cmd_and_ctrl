package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cosmic Epiphany — Sorcery {4}{U}{U} (Dominaria United, #5487):
//
//	"Draw cards equal to the number of instant and sorcery cards in your
//	 graveyard."
//
// The count is read as the spell resolves. Cosmic Epiphany is still on the
// stack then (CR 608.2n puts it in the graveyard last), so it is not
// counted. It is Haughty Djinn's count of the controller's instant and
// sorcery cards.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fad32e19-2eda-46b7-8d3c-b95a7dbeabc2",
		Name:         "Cosmic Epiphany",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			n := cmdrB17InstantSorceryCardsInGraveyard(ctx.Game, ctx.Controller())
			return DrawCards{Player: ctx.Controller(), N: n}.Apply(ctx)
		},
	})
}
