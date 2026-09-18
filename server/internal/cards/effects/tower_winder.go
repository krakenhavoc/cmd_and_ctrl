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
// DECLARED SIMPLIFICATION (weaker than printed): the search looks at
// your LIBRARY only. SearchLibrary is a library-zone primitive and
// the engine has no combined library-and-graveyard search yet, so a
// Command Tower already in the graveyard cannot be found. A card in
// hand is strictly better than a card in the graveyard for every
// deck that plays this, and the library is where the Tower is in
// almost every game it matters, so the loss is small and it is never
// a gain: the printed card can find strictly more places than this
// one does.
func init() {
	Register(Spec{
		OracleID:     "c1aecc8a-db3a-4158-b1f1-bfe06d782482",
		Name:         "Tower Winder",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The search looks in your library only. A Command Tower already in your graveyard can't be found.",
		},
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
						Reason:    "Choose a card named Command Tower to put into your hand",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
