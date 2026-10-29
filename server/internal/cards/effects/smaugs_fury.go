package effects

import "github.com/google/uuid"

// Smaug's Fury — Instant {1}{R} (The Hobbit, HOB):
//
//	"Target creature gets +3/+0 and gains reach and first strike until
//	 end of turn."
//
// Sure Strike's two steps with reach added to the keyword grant (#4454).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "5c034266-8e63-4b43-8f9e-564e6fff6ec1", Name: "Smaug's Fury", Completeness: CompletenessFull,
		Targets: TargetCreature("target creature"), Purpose: ForTargets(BenefitUntilEOT(0, 3, 0, "reach", "first strike")),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				return BoostUntilEOT{Target: id, Power: 3, Label: "Smaug's Fury — +3/+0"}.Apply(ctx)
			}),
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				return GrantKeywordUntilEOT{Target: id, Keywords: []string{"reach", "first strike"}, Label: "Smaug's Fury — reach and first strike"}.Apply(ctx)
			}),
		),
	})
}
