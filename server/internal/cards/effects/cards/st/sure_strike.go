package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sure Strike — #4380, S76. Uses the existing catalog seams.
func init() {
	effects.Register(effects.Spec{
		OracleID: "fb694e7e-f66e-4958-b6ed-aa74bc9ac43e", Name: "Sure Strike", Completeness: effects.CompletenessFull,
		Targets: effects.TargetCreature("target creature"), Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 3, 0, "first strike")),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.BoostUntilEOT{Target: id, Power: 3, Label: "Sure Strike — +3/+0"}.Apply(ctx)
			}),
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"first strike"}, Label: "Sure Strike — first strike"}.Apply(ctx)
			}),
		),
	})
}
