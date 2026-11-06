package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blitzball Shot — Instant {1}{G} (Final Fantasy, #4462):
//
//	"Target creature gets +3/+3 and gains trample until end of turn."
//
// A layer 7c pump and a layer 6 keyword grant on the one target, both
// until end of turn, and neither if the target is gone or no longer
// legal as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c99779b7-a271-416e-bdc2-0f9ba6ceb144",
		Name:         "Blitzball Shot",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 3, 3, "trample")),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (BoostUntilEOT{Target: id, Power: 3, Toughness: 3, Label: "Blitzball Shot — +3/+3"}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Blitzball Shot — trample"}.Apply(ctx)
		},
	})
}
