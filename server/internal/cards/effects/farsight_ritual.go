package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Farsight Ritual — Instant {2}{U}{U} (Wilds of Eldraine, #4471):
//
//	"Bargain (You may sacrifice an artifact, enchantment, or token as you
//	 cast this spell.)
//	 Look at the top four cards of your library. If this spell was
//	 bargained, look at the top eight cards of your library instead. Put
//	 two of them into your hand and the rest on the bottom of your library
//	 in a random order."
//
// Bargain is the shared optional additional cost (g2Bargain, CR 601.2b),
// read back with g2WasBargained. The look is Memory Deluge's: a mandatory
// take of two from the looked-at cards, the rest to the bottom in a
// random order (takeRestBottomRandom). A library shorter than the look
// shows what it has.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "7de66ba3-6e97-4e0a-928e-7ce45dee278b",
		Name:          "Farsight Ritual",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Tutors: 2},
		OptionalCosts: []game.AdditionalCost{g2Bargain()},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			n := 4
			if g2WasBargained(ctx) {
				n = 8
			}
			player := ctx.Controller()
			return TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, n),
				Max:    2,
				Label:  "Farsight Ritual — put two of them into your hand",
				Key:    takeRestBottomRandom,
			}.Apply(ctx)
		},
	})
}
