package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Drawn from Dreams — Sorcery {2}{U}{U}:
//
//	"Look at the top seven cards of your library. Put two of them into
//	 your hand and the rest on the bottom of your library in a random
//	 order."
//
// Farsight Ritual's look: a private look at seven, an exactly-two pick for
// the hand (a shorter library offers what it has), the rest to the bottom
// in a random order (takeRestBottomRandom).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "77acdc13-e1b5-4a6c-9be1-b987e8256f10",
		Name:         "Drawn from Dreams",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			player := ctx.Controller()
			return TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 7),
				Max:    2,
				Label:  "Drawn from Dreams — put two of them into your hand and the rest on the bottom of your library in a random order",
				Key:    takeRestBottomRandom,
			}.Apply(ctx)
		},
	})
}
