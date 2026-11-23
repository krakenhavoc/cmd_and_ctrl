package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blink of an Eye — {1}{U} Instant:
//
//	"Kicker {1}{U}
//	Return target nonland permanent to its owner's hand. If this spell was
//	kicked, draw a card."
func init() {
	effects.Register(effects.Spec{
		OracleID:      "51acd176-e84d-42c2-acb2-cceeb28a9fec",
		Name:          "Blink of an Eye",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{1}{U}")},
		Targets:       effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			draws := 0
			if ctx.WasKicked() {
				draws = 1
			}
			return effects.BounceToHand{Target: t.ID, ThenTail: effects.WithN(ctx.Tail(effects.ThenDrawCards), draws)}.Apply(ctx)
		},
	})
}
