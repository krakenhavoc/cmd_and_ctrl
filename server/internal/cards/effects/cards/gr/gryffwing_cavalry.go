package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:        "8e5e6f1d-4945-4074-a33e-c0e8b6b24a47",
		Name:            "Gryffwing Cavalry",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Training("Gryffwing Cavalry — training: put a +1/+1 counter on it"),
			effects.Targeting(effects.WheneverThisAttacks("Gryffwing Cavalry — you may pay {1}{W}: target attacking creature without flying gains flying",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					target := effects.FirstLegalCardTarget(ctx)
					if target == uuid.Nil {
						return nil
					}
					return effects.MayPay{
						Chooser:  ctx.Controller(),
						Cost:     "{1}{W}",
						Question: "Gryffwing Cavalry — pay {1}{W} to give the target attacking creature flying until end of turn?",
						Then:     gryffwingCavalryPaid,
						Carry:    []uuid.UUID{target},
					}.Apply(ctx)
				}),
				effects.TargetCreature("target attacking creature without flying", effects.AttackingCreature(), effects.WithoutKeyword("flying"))),
		},
	})
}

// gryffwingCavalryPaid is the paid half, keyed: the creature picked as the
// trigger resolved rides the prompt.
var gryffwingCavalryPaid = effects.PayUnlessBranches("pay-unless/gryffwing-cavalry-flying",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		if len(r.Carry) < 1 {
			return nil
		}
		return effects.GrantKeywordUntilEOT{Target: r.Carry[0], Keywords: []string{"flying"}, Label: "Gryffwing Cavalry — flying until end of turn"}.Apply(ctx)
	}, nil)
