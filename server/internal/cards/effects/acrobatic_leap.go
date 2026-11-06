package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Acrobatic Leap — Instant {W} (The Lost Caverns of Ixalan, #4470):
//
//	"Target creature gets +1/+3 and gains flying until end of turn.
//	 Untap it."
//
// Tethermage's Advantage with flying: the pump and the keyword are
// continuous effects locked to the target at resolution (CR 611.2c) that
// end in the cleanup step (CR 514.2), then the creature untaps
// (CR 701.26b). A target that is no longer legal gets nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1347224f-421e-4270-8e50-cf25c875909c",
		Name:         "Acrobatic Leap",
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 3, "flying")),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 3, Label: "Acrobatic Leap — +1/+3"}).Apply(ctx); err != nil {
					return err
				}
				if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"flying"}, Label: "Acrobatic Leap — flying"}).Apply(ctx); err != nil {
					return err
				}
				return UntapTarget{Target: t.ID}.Apply(ctx)
			}
			return nil
		},
	})
}
