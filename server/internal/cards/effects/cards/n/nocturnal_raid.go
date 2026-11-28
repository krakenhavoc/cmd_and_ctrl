package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nocturnal Raid — Instant {2}{B}{B} (Mirage, #5633):
//
//	"Black creatures get +2/+0 until end of turn."
//
// The set of black creatures is fixed as the spell resolves (CR 611.2c), whoever
// controls them; a creature that becomes black later is not affected.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5b4b08a5-3af8-4c33-ab5e-7bee7f79e77a",
		Name:         "Nocturnal Raid",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.OfColor(ctx.TextColor("B"))), Power: 2, Label: "Nocturnal Raid — +2/+0"}.Apply(ctx)
		},
	})
}
