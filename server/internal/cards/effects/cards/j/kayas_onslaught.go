package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kaya's Onslaught — Instant {2}{W} (Kaldheim, #5505):
//
//	"Target creature gets +1/+1 and gains double strike until end of
//	 turn.
//	 Foretell {W}"
//
// Uncaged Fury's body, plus the shared foretell special action
// (CR 702.143).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "230811bf-52e7-451e-96ec-9d57c864ac5d",
		Name:           "Kaya's Onslaught",
		Completeness:   effects.CompletenessFull,
		Purpose:        effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1, "double strike")),
		Targets:        effects.TargetCreature("target creature"),
		SpecialActions: []game.SpecialAction{effects.Foretell("{W}")},
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				if err := (effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Kaya's Onslaught — +1/+1"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"double strike"}, Label: "Kaya's Onslaught — double strike"}.Apply(ctx)
			}),
		),
	})
}
