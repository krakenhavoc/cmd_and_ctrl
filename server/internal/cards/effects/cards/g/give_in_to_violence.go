package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Give In to Violence — Instant {1}{B} (Duskmourn: House of Horror,
// #4466):
//
//	"Target creature gets +2/+2 and gains lifelink until end of turn."
//
// Adamant Will's body with lifelink in place of indestructible
// (pumpAndGrantFirstTarget): a target gone by resolution gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2bf4811f-3462-4542-acb0-101e772bfec9",
		Name:         "Give In to Violence",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2, "lifelink")),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PumpAndGrantFirstTarget(2, 2, "Give In to Violence — +2/+2 and lifelink", "lifelink")(ctx.Game, item)
		},
	})
}
