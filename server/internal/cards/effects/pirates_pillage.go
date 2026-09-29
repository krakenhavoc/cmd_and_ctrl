package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:       "7a2e866b-9642-495e-b47f-5b13a24373cc",
		Name:           "Pirate's Pillage",
		Completeness:   CompletenessFull,
		AdditionalCost: DiscardCost(1),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return drawAndTreasure(item, ctx, 2, 2)
		},
	})
}
