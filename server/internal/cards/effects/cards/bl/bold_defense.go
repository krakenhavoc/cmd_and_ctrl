package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bold Defense — Instant {2}{W} (Zendikar, S82 OLD batch 491, #5583):
//
//	"Kicker {3}{W}
//	 Creatures you control get +1/+1 until end of turn. If this spell was
//	 kicked, instead creatures you control get +2/+2 and gain first strike
//	 until end of turn."
//
// The creatures are chosen once as it resolves (CR 611.2c): +1/+1 unkicked,
// +2/+2 and first strike kicked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "d21994f2-75e1-456c-a25a-729c3269df1a",
		Name:          "Bold Defense",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{3}{W}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			mine := effects.And(effects.Creature(), effects.YouControl())
			if !ctx.WasKicked() {
				return effects.BoostUntilEOT{Match: mine, Power: 1, Toughness: 1, Label: "Bold Defense — +1/+1"}.Apply(ctx)
			}
			if err := (effects.BoostUntilEOT{Match: mine, Power: 2, Toughness: 2, Label: "Bold Defense — +2/+2"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Match: mine, Keywords: []string{"first strike"}, Label: "Bold Defense — first strike"}.Apply(ctx)
		},
	})
}
