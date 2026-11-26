package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frontier Guide — Creature — Elf Scout {1}{G}, 1/1 (Duel Decks: Zendikar vs.
// Eldrazi, #5548):
//
//	"{3}{G}, {T}: Search your library for a basic land card, put it onto the
//	 battlefield tapped, then shuffle."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5f685846-60dc-45f9-b118-02e74a443e6a",
		Name:         "Frontier Guide",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{G}, {T}: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Lands: 1},
			Cost:    effects.Plus(effects.ManaCost("{3}{G}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:        item.Controller,
					Predicate:     effects.IsBasicLand,
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Reveal:        true,
					Shuffle:       true,
					TappedOnEntry: true,
					Source:        item.SourceCardID,
					Reason:        "Frontier Guide — a basic land",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
