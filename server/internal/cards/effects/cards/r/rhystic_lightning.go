package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rhystic Lightning — Instant {2}{R} (Prophecy, #5617):
//
//	"Rhystic Lightning deals 4 damage to any target unless that
//	 permanent's controller or that player pays {2}. If they do,
//	 Rhystic Lightning deals 2 damage to the permanent or player."
//
// The target's controller (the player themself for a player target) is
// asked as the spell resolves, and the table waits for the answer (#5793,
// CR 118.12, 608.2g). Both halves are branches of one keyed handle: a
// decline deals 4, a payment deals 2. A payer who cannot fund {2} is a
// decline.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f6d371e6-e25e-4820-8cf7-f692a4d9be11",
		Name:         "Rhystic Lightning",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 0 {
				return nil
			}
			who, ok := rhysticLightningPayer(ctx, legal[0])
			if !ok {
				return nil
			}
			return effects.PayUnless{
				Chooser:  who,
				Cost:     "{2}",
				Question: "Rhystic Lightning — pay {2}, or it deals 4 damage instead of 2?",
				Else:     rhysticLightningDamage,
			}.Apply(ctx)
		},
	})
}

// rhysticLightningPayer is the player asked: the target player, or the
// target permanent's controller.
func rhysticLightningPayer(ctx *effects.Context, t game.TargetRef) (uuid.UUID, bool) {
	if t.Kind == game.TargetPlayer {
		return t.ID, true
	}
	c, ok := ctx.Game.LookupCardForEffect(t.ID)
	if !ok {
		return uuid.Nil, false
	}
	return c.Controller, true
}

// rhysticLightningDamage: 2 damage when the payment was made, 4 when it
// was not. Keyed, an on-disk identity.
var rhysticLightningDamage = effects.PayUnlessBranches("pay-unless/rhystic-lightning-damage",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error { return rhysticLightningDeal(ctx, 2) },
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error { return rhysticLightningDeal(ctx, 4) })

func rhysticLightningDeal(ctx *effects.Context, n int) error {
	for _, t := range ctx.Targets() {
		return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}.Apply(ctx)
	}
	return nil
}
