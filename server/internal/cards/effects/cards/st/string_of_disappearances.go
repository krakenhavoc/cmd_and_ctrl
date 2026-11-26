package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// String of Disappearances — Instant {U} (Modern Horizons):
//
//	"Return target creature to its owner's hand. Then that creature's
//	 controller may pay {U}{U}. If the player does, they may copy this spell
//	 and may choose a new target for that copy."
//
// The creature's controller is read as the spell resolves, before the creature
// leaves, and rides the bounce's tail as data. The payment is the controller's
// own (a pool that cannot fund {U}{U} degrades to a decline), and paying leads
// to the chain cycle's "may copy this spell" offer, the copy being controlled
// by that player with the CR 707.10c re-target (ChainOfferCopy).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "23985249-c54e-4bfa-b69a-512422a46a77",
		Name:         "String of Disappearances",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			var controller uuid.UUID
			if c, found := ctx.Game.LookupCardForEffect(t.ID); found {
				controller = c.Controller
			}
			return effects.BounceToHand{
				Target:   t.ID,
				ThenTail: effects.WithCarry(ctx.ItemTail(stringOfDisappearancesBounced), controller),
			}.Apply(ctx)
		},
	})
}

var stringOfDisappearancesBounced = effects.CardTail("then/string-of-disappearances-offer-payment", func(ctx *effects.Context, o game.TailOutcome) error {
	controller := effects.TailCarry0(o)
	if controller == uuid.Nil {
		return nil
	}
	return effects.MayPay{
		Chooser:  controller,
		Cost:     "{U}{U}",
		Question: "String of Disappearances — pay {U}{U}? (if you do, you may copy it)",
		Then:     stringOfDisappearancesPaid,
	}.Apply(ctx)
})

var stringOfDisappearancesPaid = effects.PayUnlessBranches("pay-unless/string-of-disappearances-copy",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		return effects.ChainOfferCopy(ctx.Game, ctx.Item, r.Chooser, "String of Disappearances")
	}, nil)
