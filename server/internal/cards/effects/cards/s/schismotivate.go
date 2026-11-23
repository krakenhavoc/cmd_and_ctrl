package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Schismotivate — Instant {1}{U}{R} (Guildpact, #5599):
//
//	"Target creature gets +4/+0 until end of turn. Another target
//	 creature gets -4/-0 until end of turn."
//
// Consume Strength's two clauses (the second Distinct, so "another"
// is a different creature) with power only. Each clause is re-checked
// on its own as the spell resolves (CR 608.2b): a creature that left
// in response is skipped and the other still gets its half.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d7081ec5-0070-4bc1-97c4-544ac08581ee",
		Name:         "Schismotivate",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.Distinct(effects.TargetCreature("another target creature")),
		),
		Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 4, 0), effects.HarmUntilEOT(1, -4, 0)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: 4, Label: "Schismotivate — +4/+0"}).Apply(ctx); err != nil {
					return err
				}
			}
			if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetCard {
				return effects.BoostUntilEOT{Target: t.ID, Power: -4, Label: "Schismotivate — -4/-0"}.Apply(ctx)
			}
			return nil
		},
	})
}
