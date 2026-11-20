package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d7081ec5-0070-4bc1-97c4-544ac08581ee",
		Name:         "Schismotivate",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature"),
			Distinct(TargetCreature("another target creature")),
		),
		Purpose: ForTargets(BenefitUntilEOT(0, 4, 0), HarmUntilEOT(1, -4, 0)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				if err := (BoostUntilEOT{Target: t.ID, Power: 4, Label: "Schismotivate — +4/+0"}).Apply(ctx); err != nil {
					return err
				}
			}
			if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetCard {
				return BoostUntilEOT{Target: t.ID, Power: -4, Label: "Schismotivate — -4/-0"}.Apply(ctx)
			}
			return nil
		},
	})
}
