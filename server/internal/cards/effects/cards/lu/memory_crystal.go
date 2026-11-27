package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Memory Crystal — Artifact {3} (Exodus, S82 OLD batch 745, #5626):
//
//	"Buyback costs cost {2} less."
//
// BuybackCostsLess (#6206): a reduction of the buyback component of every
// player's spells and of nothing else (CR 702.27a, 118.8, 601.2f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5fd3141-6fd9-4491-9f29-7fd81209bd57",
		Name:         "Memory Crystal",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.BuybackCostsLess(2, "Buyback costs cost {2} less."),
		},
	})
}
