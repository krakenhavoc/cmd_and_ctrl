package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seize the Initiative — Instant {W} (Scars of Mirrodin, #5577):
//
//	"Target creature gets +1/+1 and gains first strike until end of turn."
//
// Kindled Fury's shape at +1/+1 (pumpAndGrantFirstTarget).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f6653d90-cd30-4b2b-9592-97d5769dc3d8",
		Name:         "Seize the Initiative",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1, "first strike")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PumpAndGrantFirstTarget(1, 1, "Seize the Initiative — +1/+1 and first strike", "first strike")(ctx.Game, item)
		},
	})
}
