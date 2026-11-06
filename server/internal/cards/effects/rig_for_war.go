package effects

import "github.com/google/uuid"

// Rig for War — Instant {1}{R} (Edge of Eternities, #4461):
//
//	"Target creature gets +3/+0 and gains reach and first strike until
//	 end of turn."
//
// Smaug's Fury's text, word for word: the pump and the keyword grant as
// two steps on the one target, both until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "05bbd557-e2c0-4855-8fe6-00527fa39878",
		Name:         "Rig for War",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 3, 0, "reach", "first strike")),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				return BoostUntilEOT{Target: id, Power: 3, Label: "Rig for War — +3/+0"}.Apply(ctx)
			}),
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				return GrantKeywordUntilEOT{Target: id, Keywords: []string{"reach", "first strike"}, Label: "Rig for War — reach and first strike"}.Apply(ctx)
			}),
		),
	})
}
