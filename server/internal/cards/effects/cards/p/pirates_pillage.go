package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pirate's Pillage — Sorcery {3}{R}:
//
//	"As an additional cost to cast this spell, discard a card.
//	 Draw two cards and create two Treasure tokens."
//
// Big Score's shape again, one mana higher. Shares drawAndTreasure —
// unlike Seize the Spoils, this one is the SAME (2 cards, 2
// Treasures) as the helper's two existing callers, so it introduces
// no new case.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "7a2e866b-9642-495e-b47f-5b13a24373cc",
		Name:           "Pirate's Pillage",
		Purpose:        game.Purpose{Draws: 2, Tokens: 2},
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.DiscardCost(1),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DrawAndTreasure(item, ctx, 2, 2)
		},
	})
}
