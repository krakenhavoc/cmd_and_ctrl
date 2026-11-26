package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tyrannize — Sorcery {3}{B/R}{B/R} (Shadowmoor, S82 OLD b529, #5591):
//
//	"Target player discards their hand unless they pay 7 life."
//
// The targeted player is asked as the spell resolves; a player who can't pay 7
// life (below 7, CR 119.4) is not asked and discards. Paying keeps the hand,
// declining discards all of it at once (Wit's End's whole-hand discard).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "27da0af7-85d3-42aa-9426-b86c5e464f81",
		Name:         "Tyrannize",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetPlayer {
					continue
				}
				if !ctx.Game.PayActionCanPayForEffect(t.ID, effects.LifePayment(7).Action) {
					_, err := effects.DiscardWholeHand(ctx.Game, t.ID)
					return err
				}
				return effects.MayPay{
					Chooser:  t.ID,
					Pay:      effects.LifePayment(7),
					Question: "Tyrannize — pay 7 life, or discard your hand?",
					Then:     tyrannizeAnswer,
				}.Apply(ctx)
			}
			return nil
		},
	})
}

// tyrannizeAnswer: the payment itself keeps the hand; declining discards it.
// The chooser is the target, so the decline reads them off the answer. An
// on-disk key: never renamed, never reused.
var tyrannizeAnswer = effects.PayUnlessBranches("pay-unless/tyrannize-discard-hand",
	nil,
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		_, err := effects.DiscardWholeHand(ctx.Game, r.Chooser)
		return err
	})
