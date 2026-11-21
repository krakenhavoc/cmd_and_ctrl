package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Riverwise Augur — {3}{U} Creature — Merfolk Wizard 2/2:
//
//	"When this creature enters, draw three cards, then put two cards
//	 from your hand on top of your library in any order."
//
// Brainstorm on a body: the same draw-then-put-back as Brainstorm
// (brainstormOnResolve), raised when the creature enters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eac79f15-3b63-457a-b6a8-f9237439c93c",
		Name:         "Riverwise Augur",
		Purpose:      game.Purpose{Draws: 3, Answers: game.AnswerValue},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Riverwise Augur — draw three cards, then put two cards from your hand on top of your library in any order",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BrainstormOnResolve(item, effects.NewContext(g, item))
				}),
		},
	})
}
