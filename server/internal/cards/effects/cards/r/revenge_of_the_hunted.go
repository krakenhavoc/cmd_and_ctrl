package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Revenge of the Hunted — Sorcery {4}{G}{G} (Avacyn Restored, S82 OLD, #5570):
//
//	"Until end of turn, target creature gets +6/+6 and gains trample, and all
//	 creatures able to block it this turn do so.
//	 Miracle {G} (You may cast this card for its miracle cost when you draw it
//	 if it's the first card you drew this turn.)"
//
// Roar of Challenge's lure on the target, with a +6/+6 boost and trample until
// end of turn; the miracle cast is Thunderous Wrath's declaration. A target
// gone by resolution means the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "6292edfc-d75b-465e-9e10-9d259551b1a5",
		Name:             "Revenge of the Hunted",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetCreature("target creature"),
		Purpose:          effects.ForTargets(effects.BenefitUntilEOT(0, 6, 6, "trample")),
		AlternativeCosts: []game.AlternativeCost{effects.Miracle("{G}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: t.ID, Power: 6, Toughness: 6, Label: "Revenge of the Hunted — +6/+6"}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"}, Label: "Revenge of the Hunted — trample"}).Apply(ctx); err != nil {
				return err
			}
			return effects.LureTargetForTheTurn(ctx, "Revenge of the Hunted — all creatures able to block it do so")
		},
	})
}
