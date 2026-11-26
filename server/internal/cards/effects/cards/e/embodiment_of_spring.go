package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Embodiment of Spring — Creature — Elemental {U} 0/3 (Khans of Tarkir):
//
//	"{1}{G}, {T}, Sacrifice this creature: Search your library for a basic
//	 land card, put it onto the battlefield tapped, then shuffle."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e12bba04-b10f-4bbd-959c-560f31630951",
		Name:         "Embodiment of Spring",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}, {T}, Sacrifice this creature: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
			Purpose: game.Purpose{Lands: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}{G}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:        item.Controller,
					Predicate:     effects.IsBasicLand,
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Shuffle:       true,
					TappedOnEntry: true,
					Source:        item.SourceCardID,
					Reason:        "Embodiment of Spring — a basic land card",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
