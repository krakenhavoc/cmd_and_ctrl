package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Precise Strike — Instant {R} (Aether Revolt,
// #5536):
//
//	"Target creature gets +1/+0 and gains first strike until end of turn."
//
// Kindled Fury.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bfd2138b-1883-4900-a5f1-373884965c34",
		Name:         "Precise Strike",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "first strike")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PumpAndGrantFirstTarget(1, 0, "Precise Strike — +1/+0 and first strike", "first strike")(ctx.Game, item)
		},
	})
}
