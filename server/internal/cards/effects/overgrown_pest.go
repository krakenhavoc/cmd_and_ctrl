package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Overgrown Pest — Creature — Pest {2}{G}, 2/2 (March of the Machine):
//
//	"When this creature enters, look at the top five cards of your
//	 library. You may reveal a land or double-faced card from among them
//	 and put that card into your hand. Put the rest on the bottom of your
//	 library in a random order."
//
// Oath of Nissa's look with a "land or double-faced card" filter: only the
// controller sees the five, the card taken is revealed, and the rest go to
// the bottom in a random order. A double-faced card is read off the card's
// faces (CR 712.1), so a transforming, modal or meld card all count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ce84891f-07e1-4c72-a7b8-93b6afd1293e",
		Name:         "Overgrown Pest",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Overgrown Pest — look at the top five cards of your library", lookFiveTakeLandOrDoubleFaced("Overgrown Pest")),
		},
	})
}

// lookFiveTakeLandOrDoubleFaced is "look at the top five cards of your
// library. You may reveal a land or double-faced card from among them and
// put it into your hand. Put the rest on the bottom in a random order."
func lookFiveTakeLandOrDoubleFaced(name string) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		player := ctx.Controller()
		return TakeFromLibraryToHand{
			Player:   player,
			Cards:    g.LookAtTopOfLibraryForEffect(player, 5),
			Match:    Or(Land(), doubleFaced()),
			Max:      1,
			Optional: true,
			Label:    name + " — reveal a land or double-faced card and put it into your hand",
			Key:      takeRevealRestBottomRandom,
		}.Apply(ctx)
	}
}
