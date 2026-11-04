package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zimone's Experiment — Sorcery {3}{G} (Secrets of Strixhaven, #4456):
//
//	"Look at the top five cards of your library. You may reveal up to two
//	 creature and/or land cards from among them, then put the rest on the
//	 bottom of your library in a random order. Put all land cards revealed
//	 this way onto the battlefield tapped and put all creature cards revealed
//	 this way into your hand."
//
// A private look (only the controller learns the five), then one choose_cards
// prompt for up to two creature and/or land cards (CR 701.20 for the reveal).
// The chosen cards are revealed while still in the library, so every seat
// sees them; the rest stay hidden. The lands enter in one simultaneous batch,
// tapped as printed (Animist's Awakening's put), and the creature cards then
// go to the hand with the unrevealed cards put on the bottom in a random order
// (soc_b39_helpers.go). The two sentences are independent: no creature among
// the five, or none chosen, still puts the lands in and the rest on the
// bottom. With neither a creature nor a land among the five there is no
// prompt and all five go to the bottom.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5faa9e54-1cc4-48c0-bbfa-a495087c36d6",
		Name:         "Zimone's Experiment",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return socB39LookAtFiveAndSplit(ctx)
		},
	})
}
