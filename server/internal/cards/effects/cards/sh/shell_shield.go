package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shell Shield — Instant {U} (Zendikar Rising, S82 OLD #5508):
//
//	"Kicker {1} (You may pay an additional {1} as you cast this spell.)
//	 Target creature you control gets +0/+3 until end of turn. If this spell
//	 was kicked, that creature also gains hexproof until end of turn. (It
//	 can't be the target of spells or abilities your opponents control.)"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "73535033-2875-4a82-98d5-84548e349f13",
		Name:          "Shell Shield",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{1}")},
		Purpose:       effects.ForTargets(effects.BenefitUntilEOT(0, 0, 3)),
		Targets:       effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: id, Toughness: 3, Label: "Shell Shield — +0/+3"}).Apply(ctx); err != nil {
				return err
			}
			if !ctx.WasKicked() {
				return nil
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"hexproof"}, Label: "Shell Shield — hexproof"}.Apply(ctx)
		},
	})
}
