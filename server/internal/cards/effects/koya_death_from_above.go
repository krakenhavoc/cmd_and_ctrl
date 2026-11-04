package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Koya, Death from Above — Legendary Creature — Mutant Ninja Bird
// {2}{W}, 2/1 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Flying
//	 When Koya enters, exile up to one other target creature. At the
//	 beginning of the next end step, you may pay {3}{B}. If you don't,
//	 return that card to the battlefield under its owner's control."
//
// The flicker family's delayed return (exileTargetsThenScheduleReturnBody)
// with the return behind a payment: the delayed trigger is scheduled only
// when the card reached exile, at the beginning of the next end step,
// whoever's turn it is (CR 603.7), and is controlled by Koya's
// controller. As it resolves it asks that player "{3}{B}?" with
// UpkeepPayUnless, so the end step does not end while the question is
// open. Paying keeps the card in exile for good; declining, or being
// unable to pay, returns it under its owner's control as a new object. A
// card that has left exile by then stays where it is (CR 603.7c), and a
// token never comes back (CR 111.8).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e7cd87d4-028c-49d9-9416-69f57e7dfec5",
		Name:            "Koya, Death from Above",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Koya, Death from Above — exile up to one other target creature",
					func(g *game.Game, item *game.StackItem) error {
						return exileTargetsThenScheduleReturnBody(NewContext(g, item),
							"Koya, Death from Above — you may pay {3}{B}; if you don't, return the exiled card", koyaPayOrReturnBody)
					}),
				Another(TargetCreature("up to one other target creature").WithCount(0, 1))),
		},
	})
}

// koyaPayOrReturnBody is the delayed trigger: ask its controller for
// {3}{B}, and return the exiled card unless they pay.
var koyaPayOrReturnBody = game.SimpleDelayedBody("koya-death-from-above/pay-or-return",
	func(g *game.Game, item *game.StackItem) error {
		var ids []uuid.UUID
		for _, t := range item.Targets {
			if t.Kind == game.TargetCard && t.ID != uuid.Nil {
				ids = append(ids, t.ID)
			}
		}
		if len(ids) == 0 {
			return nil
		}
		return UpkeepPayUnless{
			Chooser:  item.Controller,
			Cost:     "{3}{B}",
			Question: "Koya, Death from Above — pay {3}{B}, or the exiled card returns",
			Else:     koyaReturn,
			Carry:    ids,
		}.Apply(NewContext(g, item))
	})

// koyaReturn is "if you don't, return that card to the battlefield under
// its owner's control". Keyed (#4264).
var koyaReturn = PayUnlessElse("pay-unless/koya-death-from-above-return",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		return ReturnFromExileTogether{Targets: r.Carry}.Apply(ctx)
	})
