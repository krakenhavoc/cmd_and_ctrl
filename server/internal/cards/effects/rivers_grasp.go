package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:       "f671583f-bd22-4d3e-bf19-fb740c607f8f",
		Name:           "River's Grasp",
		ReadsManaSpent: game.ReadsOwnPayment,
		Completeness:   CompletenessFull,
		WantsColors:    []string{"U", "B"},
		Targets: TargetCreature("up to one target creature").WithCount(0, 1).
			Then(TargetPlayer("target player")),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if ctx.ManaSpentOfColor("U") > 0 {
				if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
					return BounceToHand{Target: t.ID, Then: func(ctx *Context, _ bool) error {
						return riversGraspBlackHalf(ctx)
					}}.Apply(ctx)
				}
			}
			return riversGraspBlackHalf(ctx)
		},
	})
}

// riversGraspBlackHalf is the second sentence. It runs after the bounce
// has finished (a commander's owner may first be asked about the
// command zone, CR 903.9b), so a creature returned to the targeted player's hand
// is in the hand that is revealed, as the printed order has it
// (CR 608.2c).
func riversGraspBlackHalf(ctx *Context) error {
	if ctx.ManaSpentOfColor("B") == 0 {
		return nil
	}
	t, ok := ctx.ClauseTarget(1)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	return ChooseFromRevealedHand{Player: t.ID, Filter: Nonland(), Label: "nonland card"}.Apply(ctx)
}
