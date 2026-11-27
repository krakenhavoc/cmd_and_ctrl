package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hapless Researcher — Creature — Human Wizard {U}, 1/1 (Judgment):
//
//	"Sacrifice this creature: Draw a card, then discard a card."
//
// Broadside Barrage's loot (the discard opens once the draw is done, so the
// drawn card may be discarded) on a sacrifice cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "240a349a-79f8-4e11-9a05-a78148e70a62",
		Name:         "Hapless Researcher",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Discards: 1},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Draw a card, then discard a card.",
			Purpose: game.Purpose{Draws: 1, Discards: 1, Answers: game.AnswerValue},
			Cost:    effects.SacrificeThis(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.LootOne(g, item, 1)
			},
		}},
	})
}
