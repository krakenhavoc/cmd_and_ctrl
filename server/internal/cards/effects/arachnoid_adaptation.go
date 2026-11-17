package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arachnoid Adaptation — Instant {G} (March of the Machine, S82 #5477):
//
//	"Target creature gets +2/+2 and gains reach until end of turn. Untap it."
//
// The pump and the keyword are continuous effects locked to the target at
// resolution (CR 611.2c) that end in the cleanup step (CR 514.2); then the
// creature untaps (CR 701.26b). A target that is no longer legal gets
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "042cc9d9-ad13-401f-b2f5-c2f1558d7e44",
		Name:         "Arachnoid Adaptation",
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2, "reach")),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Arachnoid Adaptation — +2/+2"}).Apply(ctx); err != nil {
					return err
				}
				if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"reach"}, Label: "Arachnoid Adaptation — reach"}).Apply(ctx); err != nil {
					return err
				}
				return UntapTarget{Target: t.ID}.Apply(ctx)
			}
			return nil
		},
	})
}
