package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:        "3726b9a0-67cb-4f45-ab50-62db50f93568",
		Name:            "Ty Lee, Artful Acrobat",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Ty Lee, Artful Acrobat — you may pay {1}", func(g *game.Game, item *game.StackItem) error {
				return effects.MayPay{
					Chooser:  item.Controller,
					Cost:     "{1}",
					Question: "Ty Lee, Artful Acrobat — pay {1} to have target creature not block this turn?",
					Then:     tyLeePaid,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

var tyLeePaid = effects.PayUnlessBranches("pay-unless/ty-lee-artful-acrobat",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.WhenYouDo("Ty Lee, Artful Acrobat — target creature can't block this turn", tyLeeCantBlockBody).Apply(ctx)
	}, nil)

var tyLeeCantBlockBody = game.ReflexiveBody("ty-lee-artful-acrobat/cant-block",
	effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		target := effects.FirstLegalBattlefieldTarget(ctx)
		if target == uuid.Nil {
			return nil
		}
		return effects.RestrictUntilEOT{Target: target, Restrictions: game.CantBlock,
			Label: "Ty Lee, Artful Acrobat — can't block this turn"}.Apply(ctx)
	}),
	effects.ConstTargets(func() *game.TargetSpec { return effects.TargetCreature("target creature") }))
