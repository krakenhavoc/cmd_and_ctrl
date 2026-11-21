package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gaea's Gift — Instant for {1}{G} (The Brothers' War):
//
//	"Put a +1/+1 counter on target creature you control. It gains reach,
//	trample, hexproof, and indestructible until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2cb95721-56c8-4008-8055-579a5a97fc8b",
		Name:         "Gaea's Gift",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "reach", "trample", "hexproof", "indestructible")),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{
				Target:   t.ID,
				Keywords: []string{"reach", "trample", "hexproof", "indestructible"},
				Label:    "Gaea's Gift — reach, trample, hexproof and indestructible until end of turn",
			}.Apply(ctx)
		},
	})
}
