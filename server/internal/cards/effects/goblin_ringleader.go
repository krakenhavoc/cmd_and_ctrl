package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Ringleader — Creature — Goblin {3}{R}, 2/2 (EDHREC rank
// 4349):
//
//	"Haste (This creature can attack and {T} as soon as it comes under
//	 your control.)
//	 When this creature enters, reveal the top four cards of your
//	 library. Put all Goblin cards revealed this way into your hand and
//	 the rest on the bottom of your library in any order."
//
// The Goblin deck's refuel button, and the card that makes a Goblin
// deck a deck rather than a pile: four mana for an average of one and
// a half cards plus a hasty body, and a chain of Ringleaders that
// finds more Ringleaders.
//
// "All GOBLIN CARDS", not "all Goblin creature cards", so a Kindred
// card naming Goblin counts — which is why the filter is a bare
// subtype read. Cards in a library have no layer cache, so the printed
// type line is what is read, which is also the right answer: nothing
// on the battlefield can make a card in a library a Goblin.
//
// The take is a move out of the library, not a library SEARCH. That
// distinction matters: a search would shuffle, would prompt over the
// whole library, and would let the player pick a Goblin the reveal
// never showed them. The reveal makes all four cards known to every
// seat, as printed, and the rest go back under.
//
// Declared simplification, weaker than printed (#259): the rest go to
// the bottom in a RANDOM order rather than an order you choose. The
// engine has no ordering prompt for a pile headed to the bottom of a
// library, and random is the strictly-less-informed version of the
// choice the card gives you. It costs you the ability to set up your
// next few draws with a Ringleader that whiffed.
func init() {
	Register(Spec{
		OracleID:     "4100e486-0d27-436c-8429-76bc2c1a26ab",
		Name:         "Goblin Ringleader",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The revealed cards that aren't Goblins go to the bottom of your library in a random order — you don't get to choose the order.",
		},
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Goblin Ringleader — reveal four, take the Goblins",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return b41RevealTopThenTakeMatching(ctx, item.Controller, 4,
						func(c game.Card) bool { return c.HasSubtype("Goblin") },
						"Goblin Ringleader — the top four cards of your library")
				}),
		},
	})
}
