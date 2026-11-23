package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// River's Grasp — Sorcery {3}{U/B}:
//
//	"If {U} was spent to cast this spell, return up to one target
//	 creature to its owner's hand. If {B} was spent to cast this spell,
//	 target player reveals their hand, you choose a nonland card from
//	 it, then that player discards that card. (Do both if {U}{B} was
//	 spent.)"
//
// Moonhold's shape: each half reads the colours of mana the cast
// recorded (Context.ManaSpentOfColor, #761), over the whole cost and
// not only the hybrid symbol, and only whether any was spent (the
// 2016-06-08 rulings). Both targets are announced with the cast
// (CR 601.2c); each half acts on its own clause's pick only while it is
// still legal (CR 608.2b). The black half is the revealed-hand pick
// (ADR 0116) with Thoughtseize's nonland filter.
//
// WantsColors names both halves' colours (#3853): an auto-paid cast pays
// the generic half with a blue and a black source when the board has
// them, so "do both" is what auto-pay reaches for.
//
// A copy has no mana spent on it (the rulings), so neither half happens.
// A waived payment (strict mana off with too little in the pool, or Cast
// anyway) asks the caster what they paid outside the pool, because the
// card declares ReadsManaSpent, so the colours are exact (#3855, ADR 0068
// amendment 2026-10-15).
func init() {
	effects.Register(effects.Spec{
		OracleID:       "f671583f-bd22-4d3e-bf19-fb740c607f8f",
		Name:           "River's Grasp",
		ReadsManaSpent: game.ReadsOwnPayment,
		Completeness:   effects.CompletenessFull,
		WantsColors:    []string{"U", "B"},
		Targets: effects.TargetCreature("up to one target creature").WithCount(0, 1).
			Then(effects.TargetPlayer("target player")),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			// The mana spent is read as the spell resolves and rides the
			// bounce's keyed tail as N (#4253 slice 10a): a tail's context
			// carries the item's targets and not its payment.
			black := 0
			if ctx.ManaSpentOfColor("B") > 0 {
				black = 1
			}
			if ctx.ManaSpentOfColor("U") > 0 {
				if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
					return effects.BounceToHand{Target: t.ID, ThenTail: effects.WithN(ctx.ItemTail(riversGraspBlack), black)}.Apply(ctx)
				}
			}
			return riversGraspBlackIf(ctx, black)
		},
	})
}

// riversGraspBlackHalf is the second sentence. It runs after the bounce
// has finished (a commander's owner may first be asked about the
// command zone, CR 903.9b), so a creature returned to the targeted player's hand
// is in the hand that is revealed, as the printed order has it
// (CR 608.2c).
func riversGraspBlackHalf(ctx *effects.Context) error {
	t, ok := ctx.ClauseTarget(1)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	return effects.ChooseFromRevealedHand{Player: t.ID, Filter: effects.Nonland(), Label: "nonland card"}.Apply(ctx)
}

// riversGraspBlack is the black half after the bounce, keyed (#4253 slice
// 10a): N is 1 when {B} was spent.
var riversGraspBlack = effects.CardTail("then/rivers-grasp-black-half", func(ctx *effects.Context, o game.TailOutcome) error {
	return riversGraspBlackIf(ctx, o.Tail.N)
})

// riversGraspBlackIf is the black half when black is 1.
func riversGraspBlackIf(ctx *effects.Context, black int) error {
	if black == 0 {
		return nil
	}
	return riversGraspBlackHalf(ctx)
}
