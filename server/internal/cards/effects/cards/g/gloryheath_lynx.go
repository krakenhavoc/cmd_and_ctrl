package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gloryheath Lynx — Creature — Cat Mount {1}{W}:
//
//	"Lifelink
//	 Whenever this creature attacks while saddled, search your library
//	 for a basic Plains card, reveal it, put it into your hand, then
//	 shuffle.
//	 Saddle 2"
//
// "A basic Plains card" admits a Snow-Covered Plains and not a nonbasic
// Plains (IsBasicLandWithSubtype reads the Basic supertype). Not a
// "may": the search is mandatory, though it can find nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "28673657-f05f-4b0a-b80c-939773abdefb",
		Name:            "Gloryheath Lynx",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Activated:       []effects.ActivatedAbility{effects.Saddle(2)},
		Triggered: []game.TriggeredAbility{
			effects.AttacksWhileSaddled("Gloryheath Lynx — search for a basic Plains", func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.IsBasicLandWithSubtype("plains"),
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Gloryheath Lynx — a basic Plains card",
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
