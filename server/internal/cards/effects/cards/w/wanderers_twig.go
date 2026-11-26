package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wanderer's Twig — Artifact {1} (Lorwyn, S82 OLD b546, #5593):
//
//	"{1}, Sacrifice this artifact: Search your library for a basic land
//	 card, reveal it, put it into your hand, then shuffle."
//
// World Map's first ability without the tap: the sacrifice is a cost, and
// the search is the ability's own, so it still happens if the Twig is gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fd15a4b0-abd7-4bae-8331-81b85c94fce8",
		Name:         "Wanderer's Twig",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice this artifact: Search your library for a basic land card, reveal it, put it into your hand, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Wanderer's Twig — a basic land card, to your hand",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
