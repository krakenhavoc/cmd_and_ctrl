package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aliban's Tower — Instant {1}{R}, (Homelands, S82 OLD b823, #5635):
//
//	"Target blocking creature gets +3/+1 until end of turn."
//
// Righteousness's shape with +3/+1.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3d169bb1-e91b-46d8-9280-797192926314",
		Name:         "Aliban's Tower",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 1)),
		Targets:      effects.TargetCreature("target blocking creature", effects.BlockingCreature()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			return effects.BoostUntilEOT{Target: t.ID, Power: 3, Toughness: 1, Label: "Aliban's Tower — +3/+1"}.Apply(ctx)
		},
	})
}
