package effects

import (
	"fmt"

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
	// InThisStep holds the table in the current step until the chooser
	// answers (Game.QueueMayPayInThisStepForEffect). Set it when OnPay
	// is about the step or phase in progress — "after this phase, there
	// is an additional combat phase" (Hellkite Charger) — so the answer
	// cannot arrive after that phase has ended.
	InThisStep bool

	// Then is OnPay as REGISTERED branches (PayUnlessBranches, #4264),
	// and Carry the plain data it is handed: a prompt queued with one
	// holds no closure, so a table waiting on the payer is a restore
	// point. The key's pay branch runs only when the mana was really
	// paid. Set Then or OnPay, not both.
	Then  game.PayUnlessElse
	Carry []uuid.UUID
	// CarryObjects are objects, with the epoch each had, the card names as
	// the offer is made (a permanent it will act on): a branch is handed
	// them as PayUnlessAnswer.Objects and can tell the object from a new
	// one the same card has become (CR 400.7).
	CarryObjects []game.ObjectRef
}

func (p MayPay) Apply(ctx *Context) error {
	if p.Then.Key() != "" {
		if p.OnPay != nil {
			return fmt.Errorf("%w: MayPay %q sets both Then and OnPay", game.ErrInvalidParam, p.Then.Key())
		}
		return ctx.Game.QueuePayUnlessPromptKeyedForEffect(game.PayUnlessKeyedPrompt{
			Chooser: p.Chooser, Source: ctx.Source(), Cost: p.Cost, Question: p.Question,
			Then: p.Then, Carry: append([]uuid.UUID{ctx.Controller()}, p.Carry...),
			InThisStep: p.InThisStep, MayPay: true, Asked: game.NewAskedItem(ctx.Item),
			Objects: p.CarryObjects,
		})
	}
	item := ctx.Item
	onPay := p.OnPay
	then := func(g *game.Game) error {
		if onPay == nil {
			return nil
		}
		return onPay(NewContext(g, item))
	}
	if p.InThisStep {
		return ctx.Game.QueueMayPayInThisStepForEffect(p.Chooser, ctx.Source(), p.Cost, p.Question, then)
	}
	return ctx.Game.QueueMayPayForEffect(p.Chooser, ctx.Source(), p.Cost, p.Question, then)
}

// mayPayDrawACard is "If you do, draw a card" for every may-pay whose paid
// half is only that (Dawn of Hope, Key to the City, Leaf-Crowned Visionary,
// Lifecrafter's Bestiary, Mentor of the Meek, Mind's Eye, Seer's Sundial),
// keyed (#4264): the controller rides the asking item, so a table waiting
// on the payer is a restore point.
var mayPayDrawACard = PayUnlessBranches("pay-unless/draw-a-card",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
	}, nil)
