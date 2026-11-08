package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Compass Gnome — Artifact Creature — Gnome {2}, 2/1 (The Lost Caverns of
// Ixalan, #4470):
//
//	"When this creature enters, you may search your library for a basic
//	 land card or Cave card, reveal it, then shuffle and put that card on
//	 top."
//
// Campus Guide's search-to-top with Scampering Surveyor's "basic land
// card or Cave card": optional, revealed, and the library is shuffled
// before the card goes on top (CR 701.23a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b64051c0-e791-4fcb-b79f-4287e1b06468",
		Name:         "Compass Gnome",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Compass Gnome — search for a basic land or Cave card and put it on top", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return IsBasicLand(c) || c.HasSubtype("Cave") },
					Dest:      game.ZoneLibrary,
					ToTop:     true,
					Limit:     1,
					Optional:  true,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Compass Gnome — a basic land card or Cave card",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
