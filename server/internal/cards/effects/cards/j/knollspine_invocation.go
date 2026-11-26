package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knollspine Invocation — Enchantment {1}{R}{R} (Shadowmoor, S82 OLD b530,
// #5591):
//
//	"{X}, Discard a card with mana value X: This enchantment deals X damage to
//	 any target."
//
// X is paid in mana and is also the mana value the discarded card must have
// (DiscardCardWithManaValueX, Kozilek's cost shape), so the engine refuses a
// card that disagrees at announce. The damage is the Invocation's, read from X
// as the ability resolves, and the target is re-checked then (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a45fd5cc-0085-4d4f-872b-b9251e8706cf",
		Name:         "Knollspine Invocation",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}, Discard a card with mana value X: This enchantment deals X damage to any target.",
			Cost:    effects.Plus(effects.ManaCost("{X}"), effects.DiscardCardWithManaValueX("a card with mana value X")),
			Targets: effects.TargetAny(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				target, ok := effects.FirstLegalTarget(ctx)
				if !ok || ctx.X() <= 0 {
					return nil
				}
				ref, ok := ctx.SourceRef()
				if !ok {
					return nil
				}
				return effects.DealDamage{SourceObject: &ref, Target: target.ID, Amount: ctx.X()}.Apply(ctx)
			},
		}},
	})
}
