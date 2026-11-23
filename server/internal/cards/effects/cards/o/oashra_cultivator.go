package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oashra Cultivator — {G} Creature — Human Druid 0/3 (Jumpstart 2022, #5948):
//
//	"{2}{G}, {T}, Sacrifice this creature: Search your library for a basic
//	 land card, put it onto the battlefield tapped, then shuffle."
//
// Rampart Architect's search, run as an activated ability; the {T} waits
// out summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c07bbe66-96ae-4fab-82b9-fc6dc4f549c3",
		Name:         "Oashra Cultivator",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}, {T}, Sacrifice this creature: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
			Purpose: game.Purpose{Lands: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{G}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:        item.Controller,
					Predicate:     effects.IsBasicLand,
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					TappedOnEntry: true,
					Shuffle:       true,
					Source:        item.SourceCardID,
					Reason:        "Oashra Cultivator — a basic land card, onto the battlefield tapped",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
