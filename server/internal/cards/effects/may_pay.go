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
// effect has finished by the time the player answers, and Then's pay
// branch runs later against a Context rebuilt from the asking item. So
// anything the card does AFTER the optional cost must live inside
// Then, not after Apply returns; a statement written after Apply
// runs before the player has decided.
//
// A "Pay" the chooser cannot fund degrades to a decline (the engine
// tries pool first, then auto-taps), so the pay branch never runs
// against unpaid mana.
type MayPay struct {
	Chooser  uuid.UUID
	Cost     string
	Question string
	// InThisStep holds the table in the current step until the chooser
	// answers. Set it when the paid branch is about the step or phase in
	// progress — "after this phase, there is an additional combat phase"
	// (Hellkite Charger) — so the answer cannot arrive after that phase
	// has ended.
	InThisStep bool

	// Then is the paid branch, REGISTERED (PayUnlessBranches, #4264), and
	// Carry the plain data it is handed: the prompt holds no closure, so
	// a table waiting on the payer is a restore point. The key's pay
	// branch runs only when the mana was really paid. Required (#4253
	// retired the closure form).
	Then  game.PayUnlessElse
	Carry []uuid.UUID
	// CarryObjects are objects, with the epoch each had, the card names as
	// the offer is made (a permanent it will act on): a branch is handed
	// them as PayUnlessAnswer.Objects and can tell the object from a new
	// one the same card has become (CR 400.7).
	CarryObjects []game.ObjectRef

	// Pay is a payment that is not mana (#4566), in place of Cost: "you
	// may sacrifice an artifact or discard a card. If you do, …" is
	// EitherPayment(SacrificePayment(…), DiscardPayment(1)), one prompt
	// whose answer names the branch (ADR 0187 §5). The pay branch runs
	// once the chosen branch has been paid, and a branch the chooser
	// can't pay is never offered (CR 118.3). Zero is a mana payment of
	// Cost.
	Pay UpkeepPayment
}

func (p MayPay) Apply(ctx *Context) error {
	if p.Then.Key() == "" {
		return fmt.Errorf("%w: MayPay with no registered branches (PayUnlessBranches)", game.ErrInvalidParam)
	}
	cost, action := p.Cost, p.Pay.Action
	if action != nil {
		cost = p.Pay.Words
		// CR 118.3: a payment the chooser can make no way at all (no
		// artifact to sacrifice and no card to discard) can't be chosen,
		// so nobody is asked and "if you do" is false. The prompt would
		// otherwise offer nothing but "Don't pay".
		if !ctx.Game.PayActionCanPayForEffect(p.Chooser, action) {
			return nil
		}
	}
	return ctx.Game.QueuePayUnlessPromptKeyedForEffect(game.PayUnlessKeyedPrompt{
		Chooser: p.Chooser, Source: ctx.Source(), Cost: cost, Question: p.Question, Action: action,
		Then: p.Then, Carry: append([]uuid.UUID{ctx.Controller()}, p.Carry...),
		InThisStep: p.InThisStep, MayPay: true, Asked: game.NewAskedItem(ctx.Item),
		Objects: p.CarryObjects,
	})
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

// SacrificeAnArtifactOrDiscard is the payment "sacrifice an artifact or
// discard a card" (#4566) — K'un-Lun Warrior, Contract Hero, Vision of
// Love — and, with nonland, "… or discard a nonland card" (Bullseye,
// Death Dealer). One either payment: the chooser names the branch, and a
// branch they can't pay (no artifact, no matching card) is not offered.
func SacrificeAnArtifactOrDiscard(nonland bool) UpkeepPayment {
	discard := DiscardPayment(1)
	if nonland {
		discard = DiscardNonlandPayment(1)
	}
	return EitherPayment(SacrificePayment(1, "artifact", "artifacts", game.PermanentQuery{Types: []string{"artifact"}}), discard)
}

// DiscardNonlandPayment is "Discard N nonland cards" as a payment
// (#4566): only a nonland card in the payer's hand can pay it.
func DiscardNonlandPayment(n int) UpkeepPayment {
	p := DiscardPayment(n)
	p.Action.Nonland = true
	p.Words = "Discard a nonland card"
	if n != 1 {
		p.Words = "Discard " + numberWord(n) + " nonland cards"
	}
	return p
}
