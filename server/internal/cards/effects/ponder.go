package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ponder — Sorcery {U}:
//
//	"Look at the top three cards of your library, then put them back
//	 in any order. You may shuffle."
//	"Draw a card."
//
// Preordain's twin, one card deeper and one card less selective: you
// see three but cannot bin any of them, so the arrangement is the
// whole decision.
//
// The draw goes in LookAtTop.Then for exactly the reason Preordain's
// goes in Scry.Then — the card left on top is the card drawn, so a
// draw written on the next line would resolve before the player had
// chosen and would also leave the prompt unanswerable, its card no
// longer in the library to put back.
//
// # Declared sandbox simplification: NO SHUFFLE OPTION
//
// "You may shuffle" is not implemented. It is a second, optional
// choice that would have to be offered after the reorder is
// submitted, and it only matters when the player hates all three
// cards — at which point the arrangement they just made is discarded
// anyway. The card is still doing its job (see three, pick your
// draw); what is missing is the escape hatch from a bad three.
//
// Landing it wants a yes/no prompt chained off the reorder answer,
// which is the "prompt after a prompt" shape the choice queue does
// not have a composition for yet.
func init() {
	Register(Spec{
		OracleID: "02090581-61aa-4348-ad57-451be8ee91c2",
		Name:     "Ponder",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller := ctx.Controller()
			return LookAtTop{
				Player: controller,
				N:      3,
				Then: func(g *game.Game) error {
					return g.DrawNForEffect(controller, 1)
				},
			}.Apply(ctx)
		},
	})
}
