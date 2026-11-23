package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Righteousness — Instant {W} (Jumpstart 2022, #5948):
//
//	"Target blocking creature gets +7/+7 until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f6b2f76-0364-415e-a6a2-e9e5bf31b745",
		Name:         "Righteousness",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 7, 7)),
		Targets:      effects.TargetCreature("target blocking creature", effects.BlockingCreature()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			return effects.BoostUntilEOT{Target: t.ID, Power: 7, Toughness: 7, Label: "Righteousness — +7/+7"}.Apply(ctx)
		},
	})
}
