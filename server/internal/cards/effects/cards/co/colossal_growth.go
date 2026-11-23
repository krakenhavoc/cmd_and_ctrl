package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:      "0611beb7-a38d-4fbf-af9b-ee2a5b9a2896",
		Name:          "Colossal Growth",
		Completeness:  effects.CompletenessFull,
		Purpose:       effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3)),
		Targets:       effects.TargetCreature("target creature"),
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{R}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			if !ctx.WasKicked() {
				return effects.BoostUntilEOT{Target: t.ID, Power: 3, Toughness: 3, Label: "Colossal Growth — +3/+3"}.Apply(ctx)
			}
			if err := (effects.BoostUntilEOT{Target: t.ID, Power: 4, Toughness: 4, Label: "Colossal Growth — +4/+4"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{
				Target:   t.ID,
				Keywords: []string{"trample", "haste"},
				Label:    "Colossal Growth — trample and haste",
			}.Apply(ctx)
		},
	})
}
