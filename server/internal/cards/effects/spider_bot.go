package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spider-Bot — Artifact Creature — Spider Robot Scout {2}, 2/1 (Marvel's
// Spider-Man, #4460):
//
//	"Reach
//	 When this creature enters, you may search your library for a basic land
//	 card, reveal it, then shuffle and put that card on top."
//
// Moon-Blessed Cleric's search: SearchLibrary with ToTop finds, reveals,
// shuffles and then puts the card on top, and Optional is the printed "you may"
// (a declined search does not look, shuffle or trigger searchers).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e4bb33b2-adde-4176-9274-7598b559aa6a",
		Name:            "Spider-Bot",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Purpose:         game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Spider-Bot — search for a basic land card", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: IsBasicLand,
					Dest:      game.ZoneLibrary,
					ToTop:     true,
					Limit:     1,
					Optional:  true,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Spider-Bot — a basic land card",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
