package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Uncaged Fury — Instant {2}{R} (Foundations Jumpstart, #5929):
//
//	"Target creature gets +1/+1 and gains double strike until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7aac51ef-ae03-4a9e-9a48-46223787f291",
		Name:         "Uncaged Fury",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1, "double strike")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				if err := (effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Uncaged Fury — +1/+1"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"double strike"}, Label: "Uncaged Fury — double strike"}.Apply(ctx)
			}),
		),
	})
}
