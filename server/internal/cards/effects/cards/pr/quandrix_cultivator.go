package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quandrix Cultivator — Creature — Turtle Druid {1}{G}{G/U}{U}, 3/4
// (Strixhaven: School of Mages, #5502):
//
//	"When this creature enters, you may search your library for a basic
//	 Forest or Island card, put it onto the battlefield, then shuffle."
//
// The search is optional (a player may decline or fail to find, and the
// shuffle still happens after a search), finds a basic land card with the
// Forest or the Island subtype, and puts it onto the battlefield untapped
// (the printed text does not say tapped).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "73a63c92-6dce-4804-bcd4-3fcc942cf270",
		Name:         "Quandrix Cultivator",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 1, LandsUntapped: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Quandrix Cultivator — you may search your library for a basic Forest or Island card, put it onto the battlefield, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player: item.Controller, Predicate: effects.IsBasicLandOfAnySubtype("Forest", "Island"),
						Dest: game.ZoneBattlefield, Limit: 1, Optional: true, Shuffle: true,
						Reason: "Quandrix Cultivator — a basic Forest or Island card, onto the battlefield",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
