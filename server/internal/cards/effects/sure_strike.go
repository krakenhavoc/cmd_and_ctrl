package effects

import "github.com/google/uuid"

// Sure Strike — #4380, S76. Uses the existing catalog seams.
func init() {
	Register(Spec{
		OracleID: "fb694e7e-f66e-4958-b6ed-aa74bc9ac43e", Name: "Sure Strike", Completeness: CompletenessFull,
		Targets: TargetCreature("target creature"), Purpose: ForTargets(BenefitUntilEOT(0, 3, 0, "first strike")),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				return BoostUntilEOT{Target: id, Power: 3, Label: "Sure Strike — +3/+0"}.Apply(ctx)
			}),
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				return GrantKeywordUntilEOT{Target: id, Keywords: []string{"first strike"}, Label: "Sure Strike — first strike"}.Apply(ctx)
			}),
		),
	})
}
