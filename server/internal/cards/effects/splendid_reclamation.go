package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Splendid Reclamation — Sorcery {3}{G} (EDHREC rank 874):
//
//	"Return all land cards from your graveyard to the battlefield
//	 tapped."
//
// The lands deck's mass recursion — every fetchland cracked, every
// land discarded to a wheel, back at once. The lands are read before
// anything moves, and each returns under its owner's control, which
// is the caster's: "your graveyard".
//
// No simplification: the lands enter together (#1867), so each sees
// the others enter (CR 603.6a), and already tapped, with the CR 614
// entry stamped EntersTapped rather than the card moved untapped and
// tapped a beat later. Nothing watching for a tap event sees one.
func init() {
	Register(Spec{
		OracleID:     "13fe5e46-77a6-45d8-ac0b-c3d740eccf86",
		Name:         "Splendid Reclamation",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b10ReturnAllLandCardsFromGraveyardTapped(ctx, ctx.Controller())
		},
	})
}
