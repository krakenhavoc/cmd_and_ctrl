package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lorthos, the Tidemaker — Legendary Creature — Octopus {5}{U}{U}{U}, 8/8
// (Marvel Universe, #4460):
//
//	"Whenever Lorthos attacks, you may pay {8}. If you do, tap up to eight
//	 target permanents. Those permanents don't untap during their
//	 controllers' next untap steps."
//
// The targets are chosen when the trigger goes on the stack (CR 603.3d),
// before anyone knows whether {8} will be paid, as printed: this is not a
// reflexive trigger. The payment is MayPay's prompt, and the paid branch
// (lorthosPaid, a registered key so a table waiting on the payer is a
// restore point) taps whichever of the chosen permanents are still legal
// targets and then makes each of them skip its controller's next untap
// step (TapAndFreeze), the freeze belonging to the permanent's controller
// when the untap step comes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d9fd4ff3-5ada-40f4-b949-6fe65624a0c4",
		Name:         "Lorthos, the Tidemaker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisAttacks("Lorthos, the Tidemaker — you may pay {8}",
					func(g *game.Game, item *game.StackItem) error {
						return MayPay{
							Chooser:  item.Controller,
							Cost:     "{8}",
							Question: "Lorthos, the Tidemaker — pay {8} to tap up to eight target permanents?",
							Then:     lorthosPaid,
						}.Apply(NewContext(g, item))
					}),
				TargetPermanent("up to eight target permanents").WithCount(0, 8)),
		},
	})
}

// lorthosPaid taps the chosen permanents that are still legal targets and
// freezes them through their controllers' next untap steps.
var lorthosPaid = PayUnlessBranches("pay-unless/lorthos-tap-and-freeze",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		var ids []uuid.UUID
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard {
				ids = append(ids, t.ID)
			}
		}
		return TapAndFreeze{Targets: ids, Label: "Lorthos, the Tidemaker"}.Apply(ctx)
	}, nil)
