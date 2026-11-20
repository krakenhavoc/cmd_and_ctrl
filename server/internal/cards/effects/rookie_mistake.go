package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rookie Mistake — Instant {U} (Core Set 2021, #5511):
//
//	"Until end of turn, target creature gets +0/+2 and another target
//	 creature gets -2/-0."
//
// Two clauses, the second Distinct from the first ("another"). Each is
// re-checked on its own as it resolves, so one that left in response
// doesn't stop the other.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e51680d1-3e6c-4d1b-9705-36f73d7328cf",
		Name:         "Rookie Mistake",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature"),
			Distinct(TargetCreature("another target creature")),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if t, ok := ctx.ClauseTarget(0); ok {
				if err := (BoostUntilEOT{Target: t.ID, Toughness: 2, Label: "Rookie Mistake — +0/+2"}).Apply(ctx); err != nil {
					return err
				}
			}
			if t, ok := ctx.ClauseTarget(1); ok {
				return BoostUntilEOT{Target: t.ID, Power: -2, Label: "Rookie Mistake — -2/-0"}.Apply(ctx)
			}
			return nil
		},
	})
}
