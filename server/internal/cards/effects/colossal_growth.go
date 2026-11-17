package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Colossal Growth — {1}{G} Instant (Dominaria United):
//
//	"Kicker {R} (You may pay an additional {R} as you cast this spell.)
//	 Target creature gets +3/+3 until end of turn. If this spell was
//	 kicked, instead that creature gets +4/+4 and gains trample and haste
//	 until end of turn."
//
// CR 702.33's Kicker, read back with ctx.WasKicked: the kicked clause
// replaces the +3/+3 with +4/+4 and adds the two keywords.
func init() {
	Register(Spec{
		OracleID:      "0611beb7-a38d-4fbf-af9b-ee2a5b9a2896",
		Name:          "Colossal Growth",
		Completeness:  CompletenessFull,
		Purpose:       ForTargets(BenefitUntilEOT(0, 3, 3)),
		Targets:       TargetCreature("target creature"),
		OptionalCosts: []game.AdditionalCost{Kicker("{R}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			if !ctx.WasKicked() {
				return BoostUntilEOT{Target: t.ID, Power: 3, Toughness: 3, Label: "Colossal Growth — +3/+3"}.Apply(ctx)
			}
			if err := (BoostUntilEOT{Target: t.ID, Power: 4, Toughness: 4, Label: "Colossal Growth — +4/+4"}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Target:   t.ID,
				Keywords: []string{"trample", "haste"},
				Label:    "Colossal Growth — trample and haste",
			}.Apply(ctx)
		},
	})
}
