package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blink of an Eye — {1}{U} Instant:
//
//	"Kicker {1}{U}
//	Return target nonland permanent to its owner's hand. If this spell was
//	kicked, draw a card."
func init() {
	Register(Spec{
		OracleID:      "51acd176-e84d-42c2-acb2-cceeb28a9fec",
		Name:          "Blink of an Eye",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Kicker("{1}{U}")},
		Targets:       TargetPermanent("target nonland permanent", Nonland()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			draws := 0
			if ctx.WasKicked() {
				draws = 1
			}
			return BounceToHand{Target: t.ID, ThenTail: withN(ctx.Tail(thenDrawCards), draws)}.Apply(ctx)
		},
	})
}
