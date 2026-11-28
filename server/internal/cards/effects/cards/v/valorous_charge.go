package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Valorous Charge — Sorcery {1}{W}{W} (Portal, S82 OLD batch 784, #5630):
//
//	"White creatures get +2/+0 until end of turn."
//
// Every white creature, whoever controls it, as the spell resolves (CR 611.2c:
// a creature that becomes white or enters later is not affected).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1f51fea2-ce19-44ef-a330-65cc0fafcd64",
		Name:         "Valorous Charge",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.OfColor("W")), Power: 2,
				Label: "Valorous Charge — white creatures get +2/+0 until end of turn",
			}.Apply(ctx)
		},
	})
}
