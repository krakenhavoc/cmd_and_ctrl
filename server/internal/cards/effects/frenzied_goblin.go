package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Frenzied Goblin — Creature — Goblin Berserker {R}, 1/1 (Foundations,
// #4465):
//
//	"Whenever this creature attacks, you may pay {R}. If you do, target
//	 creature can't block this turn."
//
// The target, any creature, is chosen as the trigger goes on the stack
// (CR 603.3d), before anyone knows whether {R} will be paid: "if you do"
// is not a reflexive trigger, so this is Lorthos, the Tidemaker's shape
// and not Ty Lee's. The {R} is MayPay's prompt as the trigger resolves;
// the paid branch (frenziedGoblinPaid, a registered key so a table
// waiting on the payer is a restore point) stops the target blocking for
// the rest of the turn if it is still a legal target (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6ac470a1-c2be-4971-a5ca-10bb189ebe4d",
		Name:         "Frenzied Goblin",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisAttacks("Frenzied Goblin — you may pay {R}", func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{R}",
						Question: "Frenzied Goblin — pay {R} so target creature can't block this turn?",
						Then:     frenziedGoblinPaid,
					}.Apply(NewContext(g, item))
				}),
				TargetCreature("target creature")),
		},
	})
}

// frenziedGoblinPaid stops the chosen creature blocking this turn, if it
// is still a legal target.
var frenziedGoblinPaid = PayUnlessBranches("pay-unless/frenzied-goblin-cant-block",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		target, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return RestrictUntilEOT{Target: target, Restrictions: game.CantBlock,
			Label: "Frenzied Goblin — can't block this turn"}.Apply(ctx)
	}, nil)
