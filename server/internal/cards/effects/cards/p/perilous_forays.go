package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Perilous Forays — Enchantment {3}{G}{G} (Ravnica Remastered):
//
//	"{1}, Sacrifice a creature: Search your library for a land card with a
//	 basic land type, put it onto the battlefield tapped, then shuffle."
//
// A land TYPE, not the basic supertype (hasABasicLandType), as Boseiju does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d552f5cb-19b6-4def-90d5-6d2d9570f93d",
		Name:         "Perilous Forays",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice a creature: Search your library for a land card with a basic land type, put it onto the battlefield tapped, then shuffle.",
			Purpose: game.Purpose{Lands: 1, Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeACreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:        item.Controller,
					Source:        item.SourceCardID,
					Predicate:     effects.HasABasicLandType,
					Dest:          game.ZoneBattlefield,
					TappedOnEntry: true,
					Limit:         1,
					Shuffle:       true,
					Reason:        "Perilous Forays — a land card with a basic land type",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
