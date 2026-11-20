package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Squadron Hawk — Creature — Bird {1}{W}, 1/1 (Magic 2011, #5579):
//
//	"Flying
//	 When this creature enters, you may search your library for up to
//	 three cards named Squadron Hawk, reveal them, put them into your
//	 hand, then shuffle."
//
// Flying is the printed keyword. The enters trigger is Avarax's named
// search with a limit of three: the search names a quality, so the
// searcher may find fewer than three (CR 701.23b), the cards found are
// revealed, and the library is shuffled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "00ab9841-934f-4a66-a98d-68d01661b1c9",
		Name:            "Squadron Hawk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Optional(WhenThisEnters("Squadron Hawk — search your library for up to three cards named Squadron Hawk",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return game.HasName(c, "Squadron Hawk") },
						Dest:      game.ZoneHand,
						Limit:     3,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Squadron Hawk — up to three cards named Squadron Hawk",
					}.Apply(NewContext(g, item))
				}),
				"Search your library for up to three cards named Squadron Hawk?"),
		},
	})
}
