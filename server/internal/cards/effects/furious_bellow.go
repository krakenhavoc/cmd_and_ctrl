package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Furious Bellow — Instant {1}{R} (Dominaria United, #5487):
//
//	"Target creature gets +3/+0 and gains first strike until end of turn.
//	 Scry 1."
//
// The pump and the keyword go on the one target, then the scry (CR 701.22a).
// A target that has become illegal counters the whole spell, scry included
// (CR 608.2b).
//
// No simplification.
func init() {
	purpose := ForTargets(BenefitUntilEOT(0, 3, 0, "first strike"))
	purpose.Scry = 1
	Register(Spec{
		OracleID:     "4fb0ff70-eba1-4485-b68d-444eb512a2a7",
		Name:         "Furious Bellow",
		Completeness: CompletenessFull,
		Purpose:      purpose,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				if err := (BoostUntilEOT{Target: t.ID, Power: 3, Label: "Furious Bellow — +3/+0"}).Apply(ctx); err != nil {
					return err
				}
				if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"first strike"}, Label: "Furious Bellow — first strike"}).Apply(ctx); err != nil {
					return err
				}
			}
			return Scry{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
