package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// may_pay.go — the optional-cost-on-a-trigger primitive. Its own
// file rather than primitives.go, per the convention enters_tapped.go
// set: concurrent card batches collide on shared files.

// MayPay is "you may pay <Cost>. If you do, <effect>" — the mirror
// image of PayUnless, sharing its prompt and its payment path.
//
// Both are one dialog with two answers and a consequence hanging off
// one of them; they differ only in WHICH answer carries the
// consequence. Rhystic Study's rider happens when the payment does
// not; Hashaton's happens when it does. Modelling the second as a
// second PendingChoice kind would have duplicated the auto-tap and
// the client dialog for no rules difference.
//
// Like PayUnless this primitive only QUEUES the prompt — the calling
// effect has finished by the time the player answers, and OnPay runs
// later against a fresh Context bound to the same stack item. So
// anything the card does AFTER the optional cost must live inside
// OnPay, not after Apply returns; a statement written after Apply
// runs before the player has decided.
//
// A "Pay" the chooser cannot fund degrades to a decline (the engine
// tries pool first, then auto-taps), so OnPay never runs against
// unpaid mana.
type MayPay struct {
	Chooser  uuid.UUID
	Cost     string
	Question string
	OnPay    func(ctx *Context) error
}

func (p MayPay) Apply(ctx *Context) error {
	item := ctx.Item
	onPay := p.OnPay
	return ctx.Game.QueueMayPayForEffect(p.Chooser, ctx.Source(), p.Cost, p.Question,
		func(g *game.Game) error {
			if onPay == nil {
				return nil
			}
			return onPay(NewContext(g, item))
		})
}
