package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Colossal Might — Instant {R}{G} (Alara Reborn, #5586):
//
//	"Target creature gets +4/+2 and gains trample until end of turn."
//
// Blitzball Shot's shape at +4/+2: a layer 7c boost and a layer 6 trample
// grant, both ending in the cleanup step. A target that left in response
// gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1d8e39c9-e04d-4328-af66-db6e1241ac06",
		Name:         "Colossal Might",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 4, 2, "trample")),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (BoostUntilEOT{Target: id, Power: 4, Toughness: 2, Label: "Colossal Might — +4/+2"}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Colossal Might — trample"}.Apply(ctx)
		},
	})
}
