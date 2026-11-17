package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Whirling Strike — Instant for {1}{R} (The Brothers' War):
//
//	"Target creature gets +2/+0 and gains first strike and trample until
//	end of turn."
func init() {
	const label = "Whirling Strike — +2/+0, first strike and trample until end of turn"
	Register(Spec{
		OracleID:     "6b37b0ef-7e89-447f-a189-b3ea7b42143f",
		Name:         "Whirling Strike",
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 0, "first strike", "trample")),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			if err := (BoostUntilEOT{Target: t.ID, Power: 2, Label: label}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"first strike", "trample"}, Label: label}.Apply(ctx)
		},
	})
}
