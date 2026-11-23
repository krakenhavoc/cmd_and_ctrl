package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phalanx Tactics — Instant {1}{W} (Jumpstart 2022, #5948):
//
//	"Target creature you control gets +2/+1 until end of turn. Each
//	 other creature you control gets +1/+1 until end of turn."
//
// The "each other creature" set is fixed as the spell resolves
// (CR 611.2c) and leaves out the target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9962a9cc-6e9f-461e-a9d3-8c28ec34caf1",
		Name:         "Phalanx Tactics",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 1)),
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 1, Label: "Phalanx Tactics — +2/+1"}).Apply(ctx); err != nil {
				return err
			}
			return effects.BoostUntilEOT{
				Match:     effects.And(effects.Creature(), effects.YouControl(), effects.OtherThan(t.ID)),
				Power:     1,
				Toughness: 1,
				Label:     "Phalanx Tactics — each other creature you control gets +1/+1",
			}.Apply(ctx)
		},
	})
}
