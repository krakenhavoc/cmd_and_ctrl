package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quirion Trailblazer — Creature — Elf Scout {3}{G}, 1/2 (Invasion, S82 #5616):
//
//	"When this creature enters, you may search your library for a basic land
//	 card, put that card onto the battlefield tapped, then shuffle."
//
// The "you may" is the controller's choice as the trigger resolves; the land
// enters tapped and the library is shuffled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a26c4076-c488-4700-8e32-4ceeeac2c29a",
		Name:         "Quirion Trailblazer",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 1},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WhenThisEnters("Quirion Trailblazer — search for a basic land card, put it onto the battlefield tapped, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:        item.Controller,
						Predicate:     effects.IsBasicLand,
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						TappedOnEntry: true,
						Shuffle:       true,
						Reason:        "Quirion Trailblazer — search for a basic land card",
						Source:        item.SourceCardID,
					}.Apply(effects.NewContext(g, item))
				}),
				"Quirion Trailblazer — search for a basic land card?"),
		},
	})
}
