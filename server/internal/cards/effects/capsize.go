package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Capsize — "Buyback {3}. Return target permanent to its owner's
// hand."
//
// Constant Mists' sibling and the MANA half of buyback (CR 702.27):
// same keyword, same engine route, an additional {3} instead of a
// land. It is here because the two halves fail differently — a
// non-mana buyback is a payment the caster picks and a mana one is a
// number added to the total at CR 601.2f — and one card cannot prove
// both.
//
// Nothing about the return is in this file. The spell bounces its
// target and then the ENGINE routes the spell itself to its owner's
// hand, because the paid record says buyback was paid (ADR 0073 §6).
// The observable difference from a card that returned itself in
// OnResolve: a Capsize whose only target left in response is
// countered by game rules, never resolves, and goes to the graveyard
// — CR 702.27b returns it "as it resolves", and that is exactly the
// case the `resolved` half of the stack-exit route exists for.
func init() {
	Register(Spec{
		OracleID:     "77637eff-2963-4402-88f3-ca346f762fc8",
		Name:         "Capsize",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target permanent"),
		OptionalCosts: []game.AdditionalCost{
			Buyback("{3}"),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return BounceToHand{Target: t.ID}.Apply(ctx)
			}
			return nil
		},
	})
}
