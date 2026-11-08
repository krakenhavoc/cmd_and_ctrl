package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Campus Guide — Artifact Creature — Golem {2}, 2/1 (Foundations,
// #4465):
//
//	"When this creature enters, you may search your library for a basic
//	 land card, reveal it, then shuffle and put that card on top."
//
// Moon-Blessed Cleric's search with a basic land for the enchantment:
// optional, revealed, shuffled, then the card goes on top.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5b8e0913-47e9-4d56-971d-abd91b0f7587",
		Name:         "Campus Guide",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Campus Guide — search for a basic land card and put it on top", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: IsBasicLand,
					Dest:      game.ZoneLibrary,
					ToTop:     true,
					Limit:     1,
					Optional:  true,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Campus Guide — a basic land card",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
