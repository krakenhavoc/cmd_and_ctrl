package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Definitely Not a Turtle — Creature — Mutant Ninja Turtle {3}, 3/2
// (Teenage Mutant Ninja Turtles Commander, #5011):
//
//	"When this creature dies, look at the top six cards of your library.
//	 You may reveal a land or legendary Turtle card from among them and
//	 put it into your hand. Put the rest on the bottom of your library in
//	 a random order."
//
// Ingenious Smith's LookAtTopThenMayTakeToHand on a dies trigger: only
// the controller sees the six, the card taken is revealed, and the rest
// go to the bottom in a random order. "A land or legendary Turtle card"
// is any land card, or a card that is both legendary and a Turtle.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3d8dfd57-7d26-4a3b-a9c3-33077c85b3db",
		Name:         "Definitely Not a Turtle",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Definitely Not a Turtle — look at the top six cards; you may reveal a land or legendary Turtle card and put it into your hand",
				LookAtTopThenMayTakeToHand(6, Or(Land(), And(Legendary(), Subtype("Turtle"))), 1,
					"Definitely Not a Turtle — reveal a land or legendary Turtle card and put it into your hand")),
		},
	})
}
