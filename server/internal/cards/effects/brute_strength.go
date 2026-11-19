package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brute Strength — {1}{R} Instant:
//
//	"Target creature gets +3/+1 and gains trample until end of turn."
func init() {
	Register(Spec{
		OracleID:     "b7a16caa-96e0-40c5-8fcb-81e06d26f92a",
		Name:         "Brute Strength",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 3, 1, "trample")),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			if err := (BoostUntilEOT{Target: t.ID, Power: 3, Toughness: 1, Label: "Brute Strength — +3/+1"}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"}, Label: "Brute Strength — trample"}.Apply(ctx)
		},
	})
}
