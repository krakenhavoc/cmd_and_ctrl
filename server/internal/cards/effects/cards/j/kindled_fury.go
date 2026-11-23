package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kindled Fury — Instant {R} (Foundations, #4465):
//
//	"Target creature gets +1/+0 and gains first strike until end of
//	 turn."
//
// Kindled Heroism without the scry: pumpAndGrantFirstTarget on the target
// still legal as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ebeee06e-3345-4f02-896f-cb0b2cfa5548",
		Name:         "Kindled Fury",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "first strike")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PumpAndGrantFirstTarget(1, 0, "Kindled Fury — +1/+0 and first strike", "first strike")(ctx.Game, item)
		},
	})
}
