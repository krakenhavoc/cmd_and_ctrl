package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scampering Surveyor — Artifact Creature — Gnome {4}, 3/2 (The Lost
// Caverns of Ixalan, #4470):
//
//	"When this creature enters, search your library for a basic land card
//	 or Cave card, put it onto the battlefield tapped, then shuffle."
//
// The basic-land search onto the battlefield tapped (the shape of
// b07SearchBasicOntoBattlefield) with "or Cave card" added to the
// predicate: a Cave is a land type, so any card with the Cave subtype
// qualifies, basic or not. The searcher may find nothing, and the library
// is shuffled either way.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "63c8f510-eba8-4f00-9f80-fbbc74801a4a",
		Name:         "Scampering Surveyor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Scampering Surveyor — search for a basic land or Cave card, onto the battlefield tapped",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:        item.Controller,
						Predicate:     func(c game.Card) bool { return effects.IsBasicLand(c) || c.HasSubtype("Cave") },
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						Shuffle:       true,
						TappedOnEntry: true,
						Reason:        "Scampering Surveyor — a basic land card or Cave card, onto the battlefield tapped",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
