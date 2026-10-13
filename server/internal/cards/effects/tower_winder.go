package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tower Winder — Creature — Snake {1}{G}, 1/1 (EDHREC rank 4513):
//
//	"Reach, deathtouch
//	 When this creature enters, search your library and/or graveyard
//	 for a card named Command Tower, reveal it, and put it into your
//	 hand. If you search your library this way, shuffle."
//
// A two-mana deathtouch blocker with reach that also fixes your
// mana. Every multicolour Commander deck plays Command Tower, so the
// Winder is effectively "draw a land" stapled to a body that trades
// up with anything in the air — and that combination is why a
// two-mana 1/1 shows up in this many lists.
//
// Reach and deathtouch ride PrintedKeywords; the search is a
// name-matched tutor to hand, revealed, with the shuffle the printed
// text requires.
//
// The search reads the library AND the graveyard (#3217,
// SearchLibrary.AlsoGraveyard): a Command Tower in the graveyard is a
// public match, but "and/or" lets the player choose their zones, so
// they may still find nothing; a Tower taken from the graveyard skips
// the shuffle ("If you search your library this way"), and one taken
// from the library, or none found, shuffles (declared simplification:
// finding nothing counts as a library search; a player who looked only
// in the graveyard and found nothing would not have shuffled, a no-op
// in play).
func init() {
	Register(Spec{
		OracleID:        "c1aecc8a-db3a-4158-b1f1-bfe06d782482",
		Name:            "Tower Winder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "deathtouch"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, b06SelfETB, "Tower Winder — search for Command Tower",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.Name == "Command Tower" },
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						// #3217: "library and/or graveyard".
						AlsoGraveyard: true,
						Reason:        "Choose a card named Command Tower to put into your hand",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
