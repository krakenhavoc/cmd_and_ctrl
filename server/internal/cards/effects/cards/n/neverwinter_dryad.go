package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Neverwinter Dryad — Creature — Dryad {G}, 1/1 (Adventures in the
// Forgotten Realms):
//
//	"{2}, Sacrifice this creature: Search your library for a basic Forest
//	 card, put it onto the battlefield tapped, then shuffle."
//
// The Landscape fetch (rfReprintALandscape) with a mana-and-sacrifice
// cost. The sacrifice is a cost, so the search happens even if nothing
// else is left of the Dryad.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a6e654fb-3567-4b92-ae9b-d064bba7112a",
		Name:         "Neverwinter Dryad",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Sacrifice this creature: Search your library for a basic Forest card, put it onto the battlefield tapped, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:        item.Controller,
					Predicate:     effects.IsBasicLandOfAnySubtype("Forest"),
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Reveal:        true,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        "Choose a basic Forest to put onto the battlefield tapped",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
