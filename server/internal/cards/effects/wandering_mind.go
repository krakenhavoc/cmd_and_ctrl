package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wandering Mind — Creature — Horror {1}{U}{R}, 2/1 (Innistrad: Crimson
// Vow, S82):
//
//	"Flying
//	 When this creature enters, look at the top six cards of your
//	 library. You may reveal a noncreature, nonland card from among them
//	 and put it into your hand. Put the rest on the bottom of your
//	 library in a random order."
//
// Adaptive Omnitool's look-six shape with a noncreature, nonland filter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0fa886c2-9b58-4368-b2d0-e619c62a554f",
		Name:            "Wandering Mind",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Answers: game.AnswerValue},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Wandering Mind — look at the top six cards, you may reveal a noncreature, nonland card and put it into your hand, the rest on the bottom in a random order",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return TakeFromLibraryToHand{
						Player:   item.Controller,
						Cards:    g.LookAtTopOfLibraryForEffect(item.Controller, 6),
						Match:    And(Noncreature(), Nonland()),
						Max:      1,
						Optional: true,
						Label:    "Wandering Mind — you may reveal a noncreature, nonland card from among them and put it into your hand",
						Key:      takeRevealRestBottomRandom,
					}.Apply(ctx)
				}),
		},
	})
}
