package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lifted by Clouds — Instant — Arcane {2}{U}:
//
//	"Target creature gains flying until end of turn.
//	Splice onto Arcane {1}{U}"
func init() {
	Register(Spec{
		OracleID:     "ef742509-4474-42d0-8bca-71ec2be3d4f1",
		Name:         "Lifted by Clouds",
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0, "flying")),
		Completeness: CompletenessFull,
		Splice:       SpliceOntoArcane("{1}{U}"),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := firstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"flying"}, Label: "Lifted by Clouds — flying"}.Apply(ctx)
		},
	})
}
