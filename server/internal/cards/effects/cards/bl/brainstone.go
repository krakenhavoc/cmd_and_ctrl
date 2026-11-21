package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brainstone — Artifact {1}:
//
//	"{2}, {T}, Sacrifice this artifact: Draw three cards, then put two
//	 cards from your hand on top of your library in any order."
//
// Brainstorm's draw-then-put-back (DrawNThenForEffect with a keyed
// DrawThen) as an activated ability. The Brainstone is sacrificed as a
// cost, so it is gone before the draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ffac1923-99d0-48e6-b1e7-168e5c324a80",
		Name:         "Brainstone",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice this artifact: Draw three cards, then put two cards from your hand on top of your library in any order.",
			Purpose: game.Purpose{Draws: 3, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return g.DrawNThenForEffect(item.Controller, 3, game.DrawThen{
					Ref: effects.Cmdr2B01BrainstonePutBack, Player: item.Controller, Source: item.SourceCardID, N: 2,
				})
			},
		}},
	})
}
