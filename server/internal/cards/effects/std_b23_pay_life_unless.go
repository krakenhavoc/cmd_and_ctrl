package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b23_pay_life_unless.go — "<chooser> may pay N life. If they don't,
// <effect>" with a chooser who is NOT the asking effect's controller
// (Meathook Massacre II's opponent's creature dying).
//
// PayUnless is the mana prompt, MayPay carries its consequence on the PAID
// answer (and runs nothing for a payer who cannot be asked), and
// AnyPlayerPays offers the payment to every player. This is the same CR
// 118.12 prompt as PayUnless over a non-mana payment: the registered Else
// runs on "no" and on a "yes" the chooser cannot fund (life can be paid only
// with that much to give, CR 119.4), exactly as PayUnless's does. Carry is
// the plain data the branch is handed after the asking controller, so a
// table waiting on the payer is a restore point.
type b23PayLifeUnless struct {
	Chooser  uuid.UUID
	Life     int
	Question string
	Else     game.PayUnlessElse
	Carry    []uuid.UUID
}

func (p b23PayLifeUnless) Apply(ctx *Context) error {
	pay := LifePayment(p.Life)
	return ctx.Game.QueuePayUnlessPromptKeyedForEffect(game.PayUnlessKeyedPrompt{
		Chooser: p.Chooser, Source: ctx.Source(), Cost: pay.Words, Question: p.Question, Action: pay.Action,
		Then: p.Else, Carry: append([]uuid.UUID{ctx.Controller()}, p.Carry...),
		Asked: game.NewAskedItem(ctx.Item),
	})
}
