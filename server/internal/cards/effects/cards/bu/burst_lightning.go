package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Burst Lightning — "Kicker {4}. Burst Lightning deals 2 damage to
// any target. If this spell was kicked, it deals 4 damage instead."
//
// The simplest possible kicker (CR 702.33) and the reason it is here:
// the whole mechanic reduces to one read at resolution. The choice
// was made at CR 601.2b, the {4} was added to the total at CR 601.2f,
// and `ctx.WasKicked()` reads the record the announcement left on the
// stack item — the same shape `ctx.PaidAltCost("overload")` gives an
// overloaded spell, and for the same reason.
//
// What it must NOT be is a recomputation. By the time this resolves
// the mana is spent and nothing on the board says whether five or one
// was paid; ADR 0073 §5 is that "was it kicked" is a fact about the
// PAYMENT, recorded when the payment happened.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ac2086fe-98ee-4280-9c7c-c5c2d6548a8b",
		Name:         "Burst Lightning",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		OptionalCosts: []game.AdditionalCost{
			effects.Kicker("{4}"),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			amount := 2
			if ctx.WasKicked() {
				amount = 4
			}
			for _, t := range ctx.LegalTargets() {
				return effects.DealDamage{
					Target: t.ID,
					Amount: amount,
					Source: item.SourceCardID,
				}.Apply(ctx)
			}
			return nil
		},
	})
}
