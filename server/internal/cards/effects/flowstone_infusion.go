package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flowstone Infusion — Instant {R} (Dominaria United, #5487):
//
//	"Target creature gets +2/-2 until end of turn."
//
// One layer 7c modification on the target. A creature whose toughness
// reaches 0 dies to the state-based check (CR 704.5f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6c75b92a-4fc1-4f26-95be-1b92f462c111",
		Name:         "Flowstone Infusion",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return BoostUntilEOT{Target: t.ID, Power: 2, Toughness: -2, Label: "Flowstone Infusion — +2/-2"}.Apply(ctx)
		},
	})
}
