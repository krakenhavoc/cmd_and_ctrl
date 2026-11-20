package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scales of Shale — Instant {2}{B} (Bloomburrow, #4467):
//
//	"Affinity for Lizards (This spell costs {1} less to cast for each
//	 Lizard you control.)
//	 Target creature gets +2/+0 and gains lifelink and indestructible
//	 until end of turn."
//
// Affinity is the shared self cost modifier over Lizards you control
// (CR 702.41a), which lowers only the generic part, so {B} is always
// paid. The rest is Shore Up's pump-and-grant on any creature, without
// the untap: a target gone by resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "655a2b78-6a6c-4b9e-b7c0-48d3768978f5",
		Name:         "Scales of Shale",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0, "lifelink", "indestructible")),
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for Lizards", effects.HasSubtype("Lizard")),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Label: "Scales of Shale — +2/+0"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"lifelink", "indestructible"},
				Label: "Scales of Shale — lifelink and indestructible"}.Apply(ctx)
		},
	})
}
