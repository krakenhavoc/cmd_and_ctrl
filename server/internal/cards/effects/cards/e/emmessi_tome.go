package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emmessi Tome — Artifact — Book {4} (Eternal Masters, S82 OLD, #5542):
//
//	"{5}, {T}: Draw two cards, then discard a card."
//
// Jalum Tome's loot with a bigger draw: two cards first, then the controller
// picks the discard from the hand that now holds them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9b7634bf-067b-4b5c-be72-ff1b80247969",
		Name:         "Emmessi Tome",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}, {T}: Draw two cards, then discard a card.",
			Purpose: game.Purpose{Draws: 2, Discards: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.B16DrawThenDiscard(g, item, 2)
			},
		}},
	})
}
