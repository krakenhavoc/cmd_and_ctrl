package game

import "github.com/google/uuid"

// pay_action_keyed.go — #4361 S8: the "yes" of a non-mana pay-unless,
// run once the payment has been made, as data.
//
// A payment that moves cards or deals damage (pay_unless_action.go) can
// stop part-way: a discarded or exiled commander asks its owner about the
// command zone (CR 903.9), and "have Browbeat deal 5 damage to you" goes
// through the CR 614 window, which can ask the payer to order two
// replacements (CR 616) or wait behind a prevention follow-up's question
// (#2040). The prompt's "yes" branch has to wait for the payment, so it
// rode the payment as a closure over the answered prompt: the discard's
// and the exile's `then`, and the damage event's damageTail.then. A table
// paused there was not a restore point.
//
// Every pay-unless is answered from its data (#4361 S3: payUnlessFrameOf
// builds its frame from the prompt), so the answered prompt's data is all
// its "yes" needs. The engine tail payUnlessPaidTail carries it: when the
// payment has been made, the tail rebuilds the prompt and runs its "yes"
// exactly as the answer would have.

// payUnlessPaid is the part of an answered pay-unless, waiting for its
// payment, that a KeyedTail has no field for: the tail carrying it
// (payUnlessPaidTailOf) holds the prompt's key as Word, its payer as
// Player, its source as Source, its carried IDs as Carry, its number as N
// and its asking item as Asked. Plain data, written with the tail.
type payUnlessPaid struct {
	// Cost is the payment's words (PayCost), which an any-player offer's
	// "paid" event is labelled with.
	Cost string `json:"cost,omitempty"`
	// Action is the payment (PayActionClause), so the frame is rebuilt
	// without parsing a cost.
	Action *PayAction `json:"action,omitempty"`
	// Offer is the any-player offer the prompt was a link of, or nil.
	Offer           *AnyPlayerOffer `json:"offer,omitempty"`
	Objects         []ObjectRef     `json:"objects,omitempty"`
	Guards          uuid.UUID       `json:"guards"`
	CauseController uuid.UUID       `json:"causeController"`
	CauseItem       uuid.UUID       `json:"causeItem"`
	AutoAnswerKey   string          `json:"autoAnswerKey,omitempty"`
}

// payUnlessPaidTailOf is the tail a non-mana payment of `c` runs once it
// has been made, or the zero tail for a prompt with no branches to run.
func payUnlessPaidTailOf(c *PendingChoice) KeyedTail {
	if c == nil || c.Kind != PendingChoicePayUnless || c.PickThen == "" {
		return KeyedTail{}
	}
	return KeyedTail{
		Ref: payUnlessPaidTail, Word: c.PickThen, Player: c.Chooser, Source: c.Source,
		Carry: copyUUIDs(c.OptionCarry), N: c.PickNumber, Asked: copyAskedItem(c.Asked),
		paid: &payUnlessPaid{
			Cost: c.PayCost, Action: copyPayAction(c.PayActionClause), Offer: cloneAnyPlayerOffer(c.AnyPlayerOffer),
			Objects: cloneObjectRefs(c.OptionObjects), Guards: c.GuardsStackItem,
			CauseController: c.PayCauseController, CauseItem: c.PayCauseItem,
			AutoAnswerKey: c.AutoAnswerKey,
		},
	}
}

// paidChoice rebuilds the answered prompt from its tail.
func paidChoice(t KeyedTail) *PendingChoice {
	p := t.paid
	return &PendingChoice{
		Kind: PendingChoicePayUnless, Chooser: t.Player, Source: t.Source, Count: 1,
		PickThen: t.Word, PayCost: p.Cost, PayActionClause: copyPayAction(p.Action),
		AnyPlayerOffer: cloneAnyPlayerOffer(p.Offer),
		OptionCarry:    copyUUIDs(t.Carry), OptionObjects: cloneObjectRefs(p.Objects), GuardsStackItem: p.Guards,
		Asked: copyAskedItem(t.Asked), PickNumber: t.N,
		PayCauseController: p.CauseController, PayCauseItem: p.CauseItem,
		AutoAnswerKey: p.AutoAnswerKey,
	}
}

// copyPayUnlessPaid deep-copies a tail's answered prompt.
func copyPayUnlessPaid(p *payUnlessPaid) *payUnlessPaid {
	if p == nil {
		return nil
	}
	out := *p
	out.Action = copyPayAction(p.Action)
	out.Offer = cloneAnyPlayerOffer(p.Offer)
	out.Objects = cloneObjectRefs(p.Objects)
	return &out
}

// unknownPayUnlessPaid is checkEffectKeys' line for an answered prompt
// whose branches (the tail's Word) this binary does not register.
func unknownPayUnlessPaid(s *keyedTailSnapshot) []string {
	switch {
	case s == nil || s.Paid == nil:
		return nil
	case s.Paid.Offer != nil && !KnownAnyPlayerPays(s.Word):
		return []string{"any-player consequences " + s.Word}
	case s.Paid.Offer == nil && !KnownPayUnlessElse(s.Word):
		return []string{"pay-unless branches " + s.Word}
	}
	return nil
}

// payUnlessPaidTail runs an answered pay-unless's "yes" once its payment
// has been made: the frame rebuilt from the prompt's data, its onPay run
// as the answered prompt's branch (answerPayUnlessLocked's runBranch).
// The key is an on-disk identity: never renamed, never reused.
var payUnlessPaidTail = RegisterTail("pay-unless/paid", func(g *Game, o TailOutcome) error {
	if o.Tail.paid == nil {
		return nil
	}
	choice := paidChoice(o.Tail)
	frame := payUnlessFrameOf(choice)
	if frame == nil {
		return nil
	}
	g.runPromptBranchLocked(choice.AutoAnswerKey, func() {
		g.runPayUnlessBranchLocked(choice, choice.Chooser, frame.onPay)
	})
	return nil
})
