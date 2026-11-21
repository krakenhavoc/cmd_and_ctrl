package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slaughter Cry — Instant {2}{R} (Zendikar, #5583):
//
//	"Target creature gets +3/+0 and gains first strike until end of
//	 turn. (It deals combat damage before creatures without first
//	 strike.)"
//
// Furious Bellow's pump without the scry: the target, if still legal,
// gets +3/+0 and first strike until end of turn (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "924e20d5-b8a0-4883-9273-97ca01bf702a",
		Name:         "Slaughter Cry",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 0, "first strike")),
		OnResolve: effects.ResolveSteps(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
			if err := (effects.BoostUntilEOT{Target: id, Power: 3, Label: "Slaughter Cry — +3/+0"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"first strike"}, Label: "Slaughter Cry — first strike"}.Apply(ctx)
		})),
	})
}
