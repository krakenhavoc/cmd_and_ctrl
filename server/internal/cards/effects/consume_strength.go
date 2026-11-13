package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Consume Strength — Instant {1}{B}{G} (Apocalypse, #5614):
//
//	"Target creature gets +2/+2 until end of turn. Another target
//	 creature gets -2/-2 until end of turn."
//
// Two clauses, the second Distinct ("another", CR 601.2c), read back by
// slot as Gurmag Rakshasa's trigger is. Each half lands only on a pick
// that is still legal as the spell resolves (CR 608.2b), so the other
// half still happens when one target has left.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "53fd9fb1-c28f-4d73-bf98-94cdab418f21",
		Name:         "Consume Strength",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature"),
			Distinct(TargetCreature("another target creature")),
		),
		Purpose: ForTargets(BenefitUntilEOT(0, 2, 2), HarmUntilEOT(1, -2, -2)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				if err := (BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Consume Strength — +2/+2"}).Apply(ctx); err != nil {
					return err
				}
			}
			if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetCard {
				return BoostUntilEOT{Target: t.ID, Power: -2, Toughness: -2, Label: "Consume Strength — -2/-2"}.Apply(ctx)
			}
			return nil
		},
	})
}
