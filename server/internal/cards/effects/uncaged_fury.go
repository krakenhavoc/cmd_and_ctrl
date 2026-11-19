package effects

import "github.com/google/uuid"

// Uncaged Fury — Instant {2}{R} (Foundations Jumpstart, #5929):
//
//	"Target creature gets +1/+1 and gains double strike until end of turn."
func init() {
	Register(Spec{
		OracleID:     "7aac51ef-ae03-4a9e-9a48-46223787f291",
		Name:         "Uncaged Fury",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 1, "double strike")),
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				if err := (BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Uncaged Fury — +1/+1"}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Target: id, Keywords: []string{"double strike"}, Label: "Uncaged Fury — double strike"}.Apply(ctx)
			}),
		),
	})
}
