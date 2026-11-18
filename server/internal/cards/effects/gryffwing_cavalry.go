package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gryffwing Cavalry — Creature — Human Knight {3}{W}, 2/2 (Crimson Vow):
//
//	"Flying
//	 Training (Whenever this creature attacks with another creature with
//	 greater power, put a +1/+1 counter on this creature.)
//	 Whenever this creature attacks, you may pay {1}{W}. If you do, target
//	 attacking creature without flying gains flying until end of turn."
//
// The target is chosen as the attack trigger goes on the stack; the {1}{W} is
// paid as it resolves (Goblin Vandal's MayPay shape) and the grant follows
// only a real payment. A target that has left or lost "attacking" by then is
// no longer legal and gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8e5e6f1d-4945-4074-a33e-c0e8b6b24a47",
		Name:            "Gryffwing Cavalry",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Training("Gryffwing Cavalry — training: put a +1/+1 counter on it"),
			Targeting(WheneverThisAttacks("Gryffwing Cavalry — you may pay {1}{W}: target attacking creature without flying gains flying",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					target := firstLegalCardTarget(ctx)
					if target == uuid.Nil {
						return nil
					}
					return MayPay{
						Chooser:  ctx.Controller(),
						Cost:     "{1}{W}",
						Question: "Gryffwing Cavalry — pay {1}{W} to give the target attacking creature flying until end of turn?",
						Then:     gryffwingCavalryPaid,
						Carry:    []uuid.UUID{target},
					}.Apply(ctx)
				}),
				TargetCreature("target attacking creature without flying", AttackingCreature(), WithoutKeyword("flying"))),
		},
	})
}

// gryffwingCavalryPaid is the paid half, keyed: the creature picked as the
// trigger resolved rides the prompt.
var gryffwingCavalryPaid = PayUnlessBranches("pay-unless/gryffwing-cavalry-flying",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if len(r.Carry) < 1 {
			return nil
		}
		return GrantKeywordUntilEOT{Target: r.Carry[0], Keywords: []string{"flying"}, Label: "Gryffwing Cavalry — flying until end of turn"}.Apply(ctx)
	}, nil)
