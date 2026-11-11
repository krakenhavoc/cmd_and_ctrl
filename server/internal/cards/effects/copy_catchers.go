package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Copy Catchers — Creature — Faerie {1}{U}, 2/1:
//
//	"Flying
//	 Whenever you surveil, you may pay {1}{U}. If you do, create a token
//	 that's a copy of this creature."
//
// The surveil trigger asks for the payment as it resolves (MayPay): the
// token is made only once the mana is paid, as a copy of the source's
// copiable values, or its last-known ones if it has left the battlefield
// (CreateTokenCopy, CR 707.2, 608.2h). Each copy has the trigger too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d8e276a8-1234-4b61-84cf-4bf231ea3129",
		Name:            "Copy Catchers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			cmdr2B27WheneverYouSurveil("Copy Catchers — you may pay {1}{U} to create a token that's a copy of this creature",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}{U}",
						Question: "Copy Catchers — pay {1}{U} to create a token that's a copy of this creature?",
						Then:     copyCatchersPaid,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// copyCatchersPaid is "create a token that's a copy of this creature", once
// {1}{U} is paid. An on-disk key: never renamed, never reused.
var copyCatchersPaid = PayUnlessBranches("pay-unless/copy-catchers-copy",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return CreateTokenCopy{Controller: ctx.Controller(), Copy: ctx.Source(), N: 1}.Apply(ctx)
	}, nil)
