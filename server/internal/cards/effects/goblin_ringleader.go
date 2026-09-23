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
// "The rest on the bottom of your library in any order" is a
// put_in_library prompt on the bottom lane (#996, ADR 0088): the player
// arranges the non-Goblins, top-first. It shipped with a random order
// and a caveat until the engine had that prompt. The revealed cards
// were seen by the whole table; the ORDER they go under in is the
// player's alone (CR 401.4).
func init() {
	Register(Spec{
		OracleID:        "4100e486-0d27-436c-8429-76bc2c1a26ab",
		Name:            "Goblin Ringleader",
		Completeness:    CompletenessFull,
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
