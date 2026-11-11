package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ranging Raptors — Creature — Dinosaur {2}{G}, 2/3 (Lost Caverns of Ixalan
// Commander):
//
//	"Enrage — Whenever this creature is dealt damage, you may search your
//	 library for a basic land card, put it onto the battlefield tapped, then
//	 shuffle."
//
// The "may" is the search's own: you find a land or none, and shuffle either
// way (CR 701.23).
func init() {
	Register(Spec{
		OracleID:     "2a86b70b-34ca-45df-8aa1-c3fb2da72dc3",
		Name:         "Ranging Raptors",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Lands: 1},
		Triggered: []game.TriggeredAbility{
			cmdr2B08EnrageTrigger("Ranging Raptors — Enrage: search for a basic land card, onto the battlefield tapped", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:        item.Controller,
					Predicate:     IsBasicLand,
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Shuffle:       true,
					TappedOnEntry: true,
					Optional:      true,
					Reason:        "Ranging Raptors — a basic land card, onto the battlefield tapped",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
