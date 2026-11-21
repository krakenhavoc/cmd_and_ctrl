package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inspired Charge — Instant {2}{W}{W}:
//
//	"Creatures you control get +2/+1 until end of turn."
//
// The affected set is fixed as it resolves (CR 611.2c): a creature that
// enters afterwards gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d465a3d6-5830-456c-8e7a-908b464db846",
		Name:         "Inspired Charge",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BoostUntilEOT{
				Match:     effects.And(effects.Creature(), effects.YouControl()),
				Power:     2,
				Toughness: 1,
				Label:     "Inspired Charge — +2/+1",
			}.Apply(ctx)
		},
	})
}
