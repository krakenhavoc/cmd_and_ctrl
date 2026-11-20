package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loam Larva — Creature — Insect {1}{G}, 1/3 (Oath of the Gatewatch, #5545):
//
//	"When this creature enters, you may search your library for a basic
//	 land card, reveal it, then shuffle and put that card on top."
//
// Campus Guide's trigger: an optional search for a basic land card,
// revealed, the library shuffled and the card put on top.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "528a19d7-ee79-45a5-b27f-e0bff4213193",
		Name:         "Loam Larva",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Loam Larva — search for a basic land card and put it on top", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: IsBasicLand,
					Dest:      game.ZoneLibrary,
					ToTop:     true,
					Limit:     1,
					Optional:  true,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Loam Larva — a basic land card",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
