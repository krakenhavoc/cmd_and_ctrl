package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Charge Through — Instant {G} (Strixhaven: School of Mages, #5502):
//
//	"Target creature gains trample until end of turn.
//	 Draw a card."
//
// Defiant Strike's shape with a keyword where the pump was: the trample,
// then the draw. A target that is gone as the spell resolves fizzles the
// whole spell, draw included (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7e24b4da-9728-43a0-8af4-f513a3e1104b",
		Name:         "Charge Through",
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0, "trample")),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Charge Through — trample"}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
