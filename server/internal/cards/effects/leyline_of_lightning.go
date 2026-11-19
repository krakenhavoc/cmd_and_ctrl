package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leyline of Lightning — Enchantment {2}{R}{R} (Enchanting Tales):
//
//	"If this card is in your opening hand, you may begin the game with it on
//	 the battlefield.
//	 Whenever you cast a spell, you may pay {1}. If you do, this enchantment
//	 deals 1 damage to target player or planeswalker."
//
// The opening-hand clause (CR 103.6a) is Spec.OpeningHand (ADR 0133). The
// trigger chooses its target as it goes on the stack (CR 603.3d) and asks for
// the {1} as it resolves (Kalastria Highborn's MayPay), so it does nothing
// when the target is no longer legal (CR 608.2b); the paid branch is the
// Leyline's own damage.
//
// No simplification.
func init() {
	t := WheneverYouCast(nil, "Leyline of Lightning — you may pay {1}: it deals 1 damage to target player or planeswalker",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			if len(ctx.LegalTargets()) == 0 {
				return nil
			}
			return MayPay{
				Chooser:  item.Controller,
				Cost:     "{1}",
				Question: "Leyline of Lightning — pay {1}? If you do, it deals 1 damage to target player or planeswalker.",
				Then:     leylineOfLightningPaid,
			}.Apply(ctx)
		})
	t.Targets = targetPlayerOrPlaneswalker()
	Register(Spec{
		OracleID:     "bb87ec08-bb1b-43d8-8809-92904cbba685",
		Name:         "Leyline of Lightning",
		Completeness: CompletenessFull,
		OpeningHand:  BeginTheGameOnTheBattlefield(),
		Triggered:    []game.TriggeredAbility{t},
	})
}

// leylineOfLightningPaid is the "if you do" half: the Leyline deals 1 damage
// to the trigger's target. An on-disk key: never renamed, never reused.
var leylineOfLightningPaid = PayUnlessBranches("pay-unless/leyline-of-lightning-ping",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return sourceDealsDamageToEachLegalTarget(1)(ctx.Game, ctx.Item)
	}, nil)
