package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Forging the Anchor — Sorcery {2}{U} (The Brothers' War, S82):
//
//	"Look at the top five cards of your library. You may reveal any number of
//	 artifact cards from among them and put the revealed cards into your hand.
//	 Put the rest on the bottom of your library in a random order."
//
// A look, then an optional, revealed take of any number of artifact cards
// (Max 0 is "any number"), and the rest in a random order on the bottom.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "17c5b5c9-5c58-46dc-b0a7-a635b9997bb8",
		Name:         "Forging the Anchor",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			player := ctx.Controller()
			return TakeFromLibraryToHand{
				Player:   player,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 5),
				Match:    Artifact(),
				Optional: true,
				Label:    "Forging the Anchor — you may reveal any number of artifact cards and put them into your hand",
				Key:      takeRevealRestBottomRandom,
			}.Apply(ctx)
		},
	})
}
