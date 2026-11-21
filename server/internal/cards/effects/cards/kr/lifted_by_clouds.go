package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lifted by Clouds — Instant — Arcane {2}{U}:
//
//	"Target creature gains flying until end of turn.
//	Splice onto Arcane {1}{U}"
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef742509-4474-42d0-8bca-71ec2be3d4f1",
		Name:         "Lifted by Clouds",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")),
		Completeness: effects.CompletenessFull,
		Splice:       effects.SpliceOntoArcane("{1}{U}"),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"flying"}, Label: "Lifted by Clouds — flying"}.Apply(ctx)
		},
	})
}
