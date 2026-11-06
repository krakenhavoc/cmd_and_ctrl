package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ty Lee, Artful Acrobat — Legendary Creature — Human Performer {2}{R},
// 3/2 (Avatar: The Last Airbender, #4459):
//
//	"Prowess
//	 Whenever Ty Lee attacks, you may pay {1}. When you do, target
//	 creature can't block this turn."
//
// Prowess is the keyword's. The attack trigger asks its controller to
// pay {1} as it resolves (Merry Bards' MayPay); paying it fires a
// reflexive trigger (CR 603.12) whose target is chosen only then, so a
// player who declines chooses nothing. The target creature can't block
// for the rest of the turn; one gone by then is not affected (CR
// 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3726b9a0-67cb-4f45-ab50-62db50f93568",
		Name:            "Ty Lee, Artful Acrobat",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Ty Lee, Artful Acrobat — you may pay {1}", func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{1}",
					Question: "Ty Lee, Artful Acrobat — pay {1} to have target creature not block this turn?",
					Then:     tyLeePaid,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}

var tyLeePaid = PayUnlessBranches("pay-unless/ty-lee-artful-acrobat",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return WhenYouDo("Ty Lee, Artful Acrobat — target creature can't block this turn", tyLeeCantBlockBody).Apply(ctx)
	}, nil)

var tyLeeCantBlockBody = game.ReflexiveBody("ty-lee-artful-acrobat/cant-block",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		target := FirstLegalBattlefieldTarget(ctx)
		if target == uuid.Nil {
			return nil
		}
		return RestrictUntilEOT{Target: target, Restrictions: game.CantBlock,
			Label: "Ty Lee, Artful Acrobat — can't block this turn"}.Apply(ctx)
	}),
	constTargets(func() *game.TargetSpec { return TargetCreature("target creature") }))
