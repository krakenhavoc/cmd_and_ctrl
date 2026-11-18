package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scattered Thoughts — Instant {3}{U} (Innistrad: Crimson Vow):
//
//	"Look at the top four cards of your library. Put two of those cards into
//	 your hand and the rest into your graveyard."
//
// Rakshasa's Bargain's look: the controller picks two of the top four (every
// card, with fewer than two in the library), and the rest are put into the
// graveyard, not milled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7594d892-1589-4643-a72a-2cfb2730a83d",
		Name:         "Scattered Thoughts",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			player := ctx.Controller()
			return TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 4),
				Max:    2,
				Label:  "Scattered Thoughts — put two of them into your hand and the rest into your graveyard",
				Key:    takeRestGraveyard,
			}.Apply(ctx)
		},
	})
}
