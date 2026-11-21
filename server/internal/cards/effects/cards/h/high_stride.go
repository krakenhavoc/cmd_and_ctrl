package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// High Stride — Instant {G} (Bloomburrow, #4467):
//
//	"Target creature gets +1/+3 and gains reach until end of turn. Untap
//	 it."
//
// Shore Up's body on any creature, with reach in place of hexproof: the
// pump and the keyword are locked to the target at resolution (CR
// 611.2c) and end in the cleanup step (CR 514.2), then it untaps (CR
// 701.26b). A target that left in response counters the spell (CR
// 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ab0aea13-1dbc-424a-ab7e-cdb303036a56",
		Name:         "High Stride",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 3, "reach")),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 3, Label: "High Stride — +1/+3"}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"reach"}, Label: "High Stride — reach"}).Apply(ctx); err != nil {
				return err
			}
			return effects.UntapTarget{Target: t.ID}.Apply(ctx)
		},
	})
}
