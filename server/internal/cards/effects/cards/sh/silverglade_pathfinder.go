package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Silverglade Pathfinder — Creature — Dryad Spellshaper {1}{G}, 1/1
// (Mercadian Masques, S82 OLD, #5619):
//
//	"{1}{G}, {T}, Discard a card: Search your library for a basic land card,
//	 put that card onto the battlefield tapped, then shuffle."
//
// The discard is a cost, so it is paid before the search whatever it finds. The
// land comes in tapped (the fetch's own clause), and the search may find nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ac00d3db-d9e4-4fc3-a4ea-dd1187f406bb",
		Name:         "Silverglade Pathfinder",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}, {T}, Discard a card: Search your library for a basic land card, put that card onto the battlefield tapped, then shuffle.",
			Purpose: game.Purpose{Lands: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}{G}"), effects.TapCost(), effects.DiscardACard()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:        item.Controller,
					Predicate:     effects.IsBasicLand,
					Dest:          game.ZoneBattlefield,
					TappedOnEntry: true,
					Limit:         1,
					Shuffle:       true,
					Reason:        "Silverglade Pathfinder — a basic land card",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
