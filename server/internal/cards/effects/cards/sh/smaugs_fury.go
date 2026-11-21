package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Smaug's Fury — Instant {1}{R} (The Hobbit, HOB):
//
//	"Target creature gets +3/+0 and gains reach and first strike until
//	 end of turn."
//
// Sure Strike's two steps with reach added to the keyword grant (#4454).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID: "5c034266-8e63-4b43-8f9e-564e6fff6ec1", Name: "Smaug's Fury", Completeness: effects.CompletenessFull,
		Targets: effects.TargetCreature("target creature"), Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 3, 0, "reach", "first strike")),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.BoostUntilEOT{Target: id, Power: 3, Label: "Smaug's Fury — +3/+0"}.Apply(ctx)
			}),
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"reach", "first strike"}, Label: "Smaug's Fury — reach and first strike"}.Apply(ctx)
			}),
		),
	})
}
